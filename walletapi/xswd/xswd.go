package xswd

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"unicode"

	"github.com/creachadair/jrpc2"
	"github.com/creachadair/jrpc2/code"
	"github.com/creachadair/jrpc2/handler"
	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/rpc"
	"github.com/deroproject/derohe/walletapi"
	"github.com/deroproject/derohe/walletapi/rpcserver"
	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

type ApplicationData struct {
	Id               string                `json:"id"`
	Name             string                `json:"name"`
	Description      string                `json:"description"`
	Url              string                `json:"url"`
	Permissions      map[string]Permission `json:"permissions"`
	Signature        []byte                `json:"signature"`
	RegisteredEvents map[rpc.EventType]bool
	// RegisteredEvents only init when accepted by user
	OnClose      chan bool     `json:"-"` // used to inform when the Session disconnect
	isRequesting bool          `json:"-"`
	limiter      *rate.Limiter `json:"-"` // rate limit requests from the application
}

func (app *ApplicationData) SetIsRequesting(value bool) {
	app.isRequesting = value
}

func (app *ApplicationData) IsRequesting() bool {
	return app.isRequesting
}

type RPCResponse struct {
	JsonRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func ResponseWithError(request *jrpc2.Request, err *jrpc2.Error) RPCResponse {
	var id string
	if request != nil {
		id = request.ID()
	}

	return RPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Error:   err,
	}
}

func ResponseWithResult(request *jrpc2.Request, result interface{}) RPCResponse {
	var id string
	if request != nil {
		id = request.ID()
	}

	return RPCResponse{
		JsonRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

type AuthorizationResponse struct {
	Message  string `json:"message"`
	Accepted bool   `json:"accepted"`
}

type Permission int

const (
	Ask Permission = iota
	Allow
	Deny
	AlwaysAllow
	AlwaysDeny
)

func (perm Permission) IsPositive() bool {
	return perm == Allow || perm == AlwaysAllow
}

func (perm Permission) String() string {
	var str string
	if perm == Ask {
		str = "Ask"
	} else if perm == Allow {
		str = "Allow"
	} else if perm == Deny {
		str = "Deny"
	} else if perm == AlwaysAllow {
		str = "Always Allow"
	} else if perm == AlwaysDeny {
		str = "Always Deny"
	} else {
		str = "Unknown"
	}

	return str
}

const PermissionDenied code.Code = -32043
const PermissionAlwaysDenied code.Code = -32044
const RateLimitExceeded code.Code = -32070

type messageRequest struct {
	app     *ApplicationData
	conn    *Connection
	request *jrpc2.Request
}

type messageRegistration struct {
	app     *ApplicationData
	conn    *Connection
	request *http.Request
}

type Connection struct {
	conn *websocket.Conn
	w    sync.Mutex
	r    sync.Mutex
}

func (c *Connection) Send(message interface{}) error {
	c.w.Lock()
	defer c.w.Unlock()
	return c.conn.WriteJSON(message)
}

func (c *Connection) Read() (int, []byte, error) {
	c.r.Lock()
	defer c.r.Unlock()
	return c.conn.ReadMessage()
}

func (c *Connection) Close() error {
	c.w.Lock()
	defer c.w.Unlock()
	return c.conn.Close()
}

type XSWD struct {
	// The websocket connected to and its app data
	applications map[*Connection]ApplicationData
	// function to request access of a dApp to wallet
	appHandler func(*ApplicationData) bool
	// function to request the permission
	requestHandler func(*ApplicationData, *jrpc2.Request) Permission
	handlerMutex   sync.Mutex
	server         *http.Server
	logger         logr.Logger
	context        *rpcserver.WalletContext
	wallet         *walletapi.Wallet_Disk
	rpcHandler     handler.Map
	running        bool
	forceAsk       bool     // forceAsk ensures no permissions can be accepted upon initial connection
	noStore        []string // noStore methods won't store AlwaysAllow permission
	requests       chan messageRequest
	registers      chan messageRegistration
	// context and cancel to cleanly exit handler_loop
	ctx    context.Context
	cancel context.CancelFunc
	// mutex for applications map
	sync.Mutex
}

// This is default port for XSWD
// It can be changed for tests only
// Production should always use 44326 as its a way to identify XSWD
const XSWD_PORT = 44326

// Create a new XSWD server which allows to connect any dApp to the wallet safely through a websocket
// Each request done by the session will wait on the appHandler and requestHandler to be accepted
// NewXSWDServer will default to forceAsk (call requestHandler) for all wallet method requests,
// methods from xswd package are default noStore and won't store AlwaysAllow permission
func NewXSWDServer(wallet *walletapi.Wallet_Disk, appHandler func(*ApplicationData) bool, requestHandler func(*ApplicationData, *jrpc2.Request) Permission) *XSWD {
	noStore := []string{"Subscribe", "SignData", "CheckSignature", "GetDaemon", "query_key", "QueryKey"}
	return NewXSWDServerWithPort(XSWD_PORT, wallet, true, noStore, appHandler, requestHandler)
}

func NewXSWDServerWithPort(port int, wallet *walletapi.Wallet_Disk, forceAsk bool, noStore []string, appHandler func(*ApplicationData) bool, requestHandler func(*ApplicationData, *jrpc2.Request) Permission) *XSWD {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("XSWD server"))
	})

	ctx, cancel := context.WithCancel(context.Background())
	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	logger := globals.Logger.WithName("XSWD")

	// Prevent crossover of custom methods to rpcserver
	xswdHandler := make(handler.Map)
	for k, v := range rpcserver.WalletHandler {
		xswdHandler[k] = v
	}

	xswd := &XSWD{
		applications:   make(map[*Connection]ApplicationData),
		appHandler:     appHandler,
		requestHandler: requestHandler,
		logger:         logger,
		server:         server,
		context:        rpcserver.NewWalletContext(logger, wallet),
		wallet:         wallet,
		// don't create a different API, we provide the same
		rpcHandler: xswdHandler,
		requests:   make(chan messageRequest),
		registers:  make(chan messageRegistration),
		running:    true,
		forceAsk:   forceAsk,
		noStore:    noStore,
		ctx:        ctx,
		cancel:     cancel,
	}

	// Register event listeners
	wallet.Wallet_Memory.AddListener(rpc.NewBalance, func(change interface{}) {
		if xswd.IsEventTracked(rpc.NewBalance) {
			xswd.BroadcastEvent(rpc.NewBalance, change)
		}
	})

	wallet.Wallet_Memory.AddListener(rpc.NewTopoheight, func(topo interface{}) {
		if xswd.IsEventTracked(rpc.NewTopoheight) {
			xswd.BroadcastEvent(rpc.NewTopoheight, topo)
		}
	})

	wallet.Wallet_Memory.AddListener(rpc.NewEntry, func(entry interface{}) {
		if xswd.IsEventTracked(rpc.NewEntry) {
			xswd.BroadcastEvent(rpc.NewEntry, entry)
		}
	})

	// Save the server in the context
	xswd.context.Extra["xswd"] = xswd

	// Register custom methods
	// HasMethod for compatibility reasons in case of custom methods declared
	xswd.SetCustomMethod("HasMethod", handler.New(HasMethod))
	xswd.SetCustomMethod("Subscribe", handler.New(Subscribe))
	xswd.SetCustomMethod("Unsubscribe", handler.New(Unsubscribe))
	xswd.SetCustomMethod("SignData", handler.New(SignData))
	xswd.SetCustomMethod("CheckSignature", handler.New(CheckSignature))
	xswd.SetCustomMethod("GetDaemon", handler.New(GetDaemon))

	mux.HandleFunc("/xswd", xswd.handleWebSocket)
	logger.Info("Starting XSWD server", "addr", server.Addr)

	go func() {
		if err := xswd.server.ListenAndServe(); err != nil {
			if xswd.running {
				logger.Error(err, "Error while starting XSWD server")
				xswd.Stop()
			}
		}
	}()

	go xswd.handler_loop()

	return xswd
}

func (x *XSWD) IsEventTracked(event rpc.EventType) bool {
	applications := x.GetApplications()
	for _, app := range applications {
		if app.RegisteredEvents[event] {
			return true
		}
	}

	return false
}

func (x *XSWD) BroadcastEvent(event rpc.EventType, value interface{}) {
	for conn, app := range x.applications {
		if app.RegisteredEvents[event] {
			if err := conn.Send(ResponseWithResult(nil, rpc.EventNotification{Event: event, Value: value})); err != nil {
				x.logger.V(2).Error(err, "Error while broadcasting event")
			}
		}
	}
}

func (x *XSWD) handler_loop() {
	for {
		select {
		case msg := <-x.requests:
			go func(msg messageRequest) {
				response := x.handleMessage(msg.app, msg.request)
				if response != nil {
					if err := msg.conn.Send(response); err != nil {
						x.logger.V(2).Error(err, "Error while writing JSON", "app", msg.app.Name)
					}
				}
			}(msg)
		case msg := <-x.registers:
			response, accepted := x.addApplication(msg.request, msg.conn, msg.app)
			if accepted {
				msg.conn.Send(AuthorizationResponse{
					Message:  response,
					Accepted: true,
				})
			} else {
				msg.conn.Send(AuthorizationResponse{
					Message:  fmt.Sprintf("Could not connect the application: %s", response),
					Accepted: false,
				})
				x.removeApplicationOfSession(msg.conn, msg.app)
			}
		case <-x.ctx.Done():
			return
		}
	}
}

func (x *XSWD) IsRunning() bool {
	return x.running
}

// Stop the XSWD server
// This will close all the connections
// and delete all applications
func (x *XSWD) Stop() {
	x.Lock()
	defer x.Unlock()
	x.running = false
	x.cancel()

	if err := x.server.Shutdown(context.Background()); err != nil {
		x.logger.Error(err, "Error while stopping XSWD server")
	}

	for conn, app := range x.applications {
		if app.IsRequesting() {
			app.OnClose <- true
		}

		conn.Close()
	}
	x.applications = make(map[*Connection]ApplicationData)
	x.logger.Info("XSWD server stopped")
	x = nil
}

// Register a custom method easily to be completely configurable
func (x *XSWD) SetCustomMethod(method string, handler handler.Func) {
	x.rpcHandler[method] = handler
}

// Get all connected Applications
// This will return a copy of the map
func (x *XSWD) GetApplications() []ApplicationData {
	x.Lock()
	defer x.Unlock()

	apps := make([]ApplicationData, 0, len(x.applications))
	for _, app := range x.applications {
		apps = append(apps, app)
	}

	return apps
}

// Remove an application
// It will automatically close the connection
func (x *XSWD) RemoveApplication(app *ApplicationData) {
	x.Lock()
	defer x.Unlock()

	for conn, a := range x.applications {
		if a.Id == app.Id {
			delete(x.applications, conn)
			if a.IsRequesting() {
				a.OnClose <- true
			}

			if err := conn.Close(); err != nil {
				x.logger.Error(err, "error while closing websocket session")
			}
			break
		}
	}
}

// Check if a application exist by its id
func (x *XSWD) HasApplicationId(app_id string) bool {
	x.Lock()
	defer x.Unlock()

	for _, a := range x.applications {
		if strings.EqualFold(a.Id, app_id) {
			return true
		}
	}
	return false
}

// Add an application from a websocket connection,
// it verifies that application is valid and will add it to the application list if user accepts the request
func (x *XSWD) addApplication(r *http.Request, conn *Connection, app *ApplicationData) (response string, accepted bool) {
	// Sanity check
	{
		id := strings.TrimSpace(app.Id)
		if len(id) != 64 {
			response = "Invalid ID size"
			x.logger.V(1).Info(response, "ID", app.Id)
			return
		}

		if _, err := hex.DecodeString(id); err != nil {
			response = "Invalid hexadecimal ID"
			x.logger.V(1).Info(response, "ID", app.Id)
			return
		}

		if len(strings.TrimSpace(app.Name)) == 0 || len(app.Name) > 255 || !isASCII(app.Name) {
			response = "Invalid name"
			x.logger.V(1).Info(response, "name", len(app.Name))
			return
		}

		if len(strings.TrimSpace(app.Description)) == 0 || len(app.Description) > 255 || !isASCII(app.Description) {
			response = "Invalid description"
			x.logger.V(1).Info(response, "description", len(app.Description))
			return
		}

		origin := r.Header.Get("Origin")
		if len(app.Url) == 0 {
			app.Url = origin
			if len(app.Url) > 0 {
				x.logger.V(1).Info("No URL passed, checking origin header")
			}
		}

		// Verify that the website url set is the same as origin (security check)
		if len(origin) > 0 && app.Url != origin {
			response = "Invalid URL compared to origin"
			x.logger.V(1).Info(response, "origin", origin, "url", app.Url)
			return
		}

		// URL can be optional
		if len(app.Url) > 255 {
			response = "Invalid URL"
			x.logger.V(1).Info(response, "url", len(app.Url))
			return
		}

		// Check that URL is starting with valid protocol
		if !(strings.HasPrefix(app.Url, "http://") || strings.HasPrefix(app.Url, "https://")) {
			response = "Invalid application URL"
			x.logger.V(1).Info(response, "url", app.Url)
			return
		}

		// Signature can be optional but if provided it must be valid for app to be added
		// and is a requirement for permissions to be set upon initial connection
		if len(app.Signature) > 0 {
			if len(app.Signature) > 512 {
				response = "Invalid signature size"
				x.logger.V(1).Info(response, "signature", len(app.Signature))
				return
			}

			signer, message, err := x.wallet.CheckSignature(app.Signature)
			if err != nil {
				response = "Invalid signature"
				x.logger.V(1).Info(response, "signature", string(app.Signature))
				return
			}

			if !signer.IsDERONetwork() {
				response = "Signer does not belong to DERO network"
				x.logger.V(1).Info(response, "signer", signer.String())
				return
			}

			// Signature message must match app ID
			mcheck := strings.TrimSpace(string(message))
			if mcheck != app.Id {
				response = "Signature does not match ID"
				x.logger.V(1).Info(response, app.Id, mcheck)
				return
			}

			x.logger.V(1).Info("Signature matches ID", app.Id, mcheck)
		} else if app.Permissions != nil && len(app.Permissions) > 0 {
			response = "Application is requesting permissions without signature"
			x.logger.V(1).Info(response, app.Name, app.Id)
			return
		}

		// Check that we don't already have this application
		if x.HasApplicationId(app.Id) {
			response = "Application ID already added"
			return
		}

		// Check permission len
		if len(app.Permissions) > 255 {
			response = "Invalid permissions"
			x.logger.V(1).Info(response, "permissions", len(app.Permissions))
			return
		}

		x.logger.Info(fmt.Sprintf("Application %s (%s) is requesting access to your wallet", app.Name, app.Url))

		// If forceAsk all permissions will default to Ask
		if !x.forceAsk {
			validPermissions := map[string]Permission{}
			normalizedMethods := map[string]Permission{}

			for n, p := range app.Permissions {
				if strings.HasPrefix(n, "DERO.") {
					x.logger.V(1).Info("Daemon requests are AlwaysAllow", n, p)
					continue
				}

				// Ensure we are not storing Allow or Deny permissions as they return positive/negative
				if p == Allow || p == Deny {
					x.logger.V(1).Info("Invalid permission requested", n, p)
					continue
				}

				// Always Ask for custom methods
				if _, ok := x.rpcHandler[n]; !ok {
					x.logger.V(1).Info("Invalid method requested", n, p)
					continue
				}

				// Check if wallet defined method as noStore
				if p == AlwaysAllow && !x.CanStorePermission(n) {
					x.logger.V(1).Info("Method not allowed AlwaysAllow permission", n, p)
					continue
				}

				// Normalize all method names
				normalized := strings.ToLower(strings.ReplaceAll(n, "_", ""))

				// Ensure if permission is added already under another method name, it matches (GetAddress == getaddress)
				if pcheck, ok := normalizedMethods[normalized]; ok && pcheck != p {
					x.logger.V(1).Info("Conflicting permissions for", n, p)
					continue
				}

				x.logger.Info("Permission requested for", n, p)
				normalizedMethods[normalized] = p
				validPermissions[n] = p
			}

			if len(validPermissions) > 0 {
				app.Permissions = validPermissions
			} else {
				x.logger.Info("All wallet requests will Ask for your permission")
				app.Permissions = map[string]Permission{}
			}
		} else {
			x.logger.Info("All wallet requests will Ask for your permission")
			app.Permissions = map[string]Permission{}
		}
	}

	// only one request at a time
	x.handlerMutex.Lock()
	defer x.handlerMutex.Unlock()

	app.OnClose = make(chan bool)
	app.limiter = rate.NewLimiter(10.0, 20)
	// check the permission from user
	app.SetIsRequesting(true)
	if x.appHandler(app) {
		app.SetIsRequesting(false)
		// check if server has stopped while in appHandler
		if !x.running {
			conn.Close()
			response = "XSWD is offline"
			x.logger.Info(response, "id", app.Id, "name", app.Name, "description", app.Description, "url", app.Url)
			return
		}

		// Create the map
		app.RegisteredEvents = map[rpc.EventType]bool{}

		x.Lock()
		x.applications[conn] = *app
		x.Unlock()

		accepted = true
		response = "User has authorized the application"
		x.logger.Info(response, "id", app.Id, "name", app.Name, "description", app.Description, "url", app.Url)
		return
	} else {
		app.SetIsRequesting(false)
		response = "User has rejected connection request"
		x.logger.Info(response, "id", app.Id, "name", app.Name, "description", app.Description, "url", app.Url)
	}

	return
}

// Remove an application from the list for a session
// only used in internal
func (x *XSWD) removeApplicationOfSession(conn *Connection, app *ApplicationData) {
	if app != nil && app.IsRequesting() {
		x.logger.Info(fmt.Sprintf("Closing %s request prompt", app.Name))
		app.OnClose <- true
	}
	conn.Close()

	x.Lock()
	vapp, found := x.applications[conn]
	delete(x.applications, conn)
	x.Unlock()

	if found {
		x.logger.Info("Application deleted", "id", vapp.Id, "name", vapp.Name, "description", vapp.Description, "url", vapp.Url)
	}
}

// Handle a RPC Request from a session
// We check that the method exists, that the application has the permission to use it
func (x *XSWD) handleMessage(app *ApplicationData, request *jrpc2.Request) interface{} {
	methodName := request.Method()
	handler := x.rpcHandler[methodName]

	// Check that the method exists
	if handler == nil {
		// Only requests methods starting with DERO. are sent to daemon
		if strings.HasPrefix(methodName, "DERO.") {
			// if daemon is online, request the daemon
			// wallet play the proxy here
			// and because no sensitive data can be obtained, we allow without requests
			if x.wallet.IsDaemonOnlineCached() {
				var params json.RawMessage
				err := request.UnmarshalParams(&params)
				if err != nil {
					x.logger.V(1).Error(err, "Error while unmarshaling params")
					return ResponseWithError(request, jrpc2.Errorf(code.InvalidParams, "Error while unmarshaling params: %q", err.Error()))
				}

				x.logger.V(2).Info("requesting daemon with", "method", request.Method(), "param", request.ParamString())
				result, err := walletapi.GetRPCClient().RPC.Call(context.Background(), request.Method(), params)
				if err != nil {
					x.logger.V(1).Error(err, "Error on daemon call")
					return ResponseWithError(request, jrpc2.Errorf(code.InvalidRequest, "Error on daemon call: %q", err.Error()))
				}

				// we set original ID
				result.SetID(request.ID())

				// Unmarshal result into response to sync wallet/daemon as RPCResponse type
				var response interface{}
				err = result.UnmarshalResult(&response)
				if err != nil {
					x.logger.V(1).Error(err, "Error on unmarshal daemon result")
					return ResponseWithError(request, jrpc2.Errorf(code.InternalError, "Error on unmarshal daemon call: %q", err.Error()))
				}

				json, err := result.MarshalJSON()
				if err != nil {
					x.logger.V(1).Error(err, "Error on marshal daemon response")
					return ResponseWithError(request, jrpc2.Errorf(code.InternalError, "Error on marshal daemon call: %q", err.Error()))
				}

				x.logger.V(2).Info("received response", "response", string(json))

				return ResponseWithResult(request, response)
			} else {
				x.logger.V(1).Info("Daemon is offline", "endpoint", x.wallet.Daemon_Endpoint)
				return ResponseWithError(request, jrpc2.Errorf(code.Cancelled, "daemon %s is offline", x.wallet.Daemon_Endpoint))
			}
		}

		x.logger.Info("RPC Method not found", "method", methodName)
		return ResponseWithError(request, jrpc2.Errorf(code.MethodNotFound, "method %q not found", methodName))
	}

	// only one request at a time
	x.handlerMutex.Lock()
	defer x.handlerMutex.Unlock()

	// check that we still have the application connected
	// otherwise don't accept as it may disconnected between both requests
	if !x.HasApplicationId(app.Id) {
		return nil
	}

	app.SetIsRequesting(true)
	perm := x.requestPermission(app, request)
	app.SetIsRequesting(false)
	if perm.IsPositive() {
		wallet_context := *x.context
		wallet_context.Extra["app_data"] = app
		ctx := context.WithValue(context.Background(), "wallet_context", &wallet_context)
		response, err := handler(ctx, request)
		if err != nil {
			return ResponseWithError(request, jrpc2.Errorf(code.InternalError, "Error while handling request method %q: %v", methodName, err))
		}

		return ResponseWithResult(request, response)
	} else {
		code := PermissionDenied
		if perm == AlwaysDeny {
			code = PermissionAlwaysDenied
		}

		x.logger.Info(fmt.Sprintf("%s permission not granted for method", app.Name), "method", methodName)
		return ResponseWithError(request, jrpc2.Errorf(code, "Permission not granted for method %q", methodName))
	}
}

// Check if method is allowed to store AlwaysAllow permission when adding application or user selection is made
func (x *XSWD) CanStorePermission(method string) bool {
	for _, m := range x.noStore {
		if m == method {
			return false
		}
	}

	return true
}

// Request the permission for a method and save its result if it must be persisted
func (x *XSWD) requestPermission(app *ApplicationData, request *jrpc2.Request) Permission {
	method := request.Method()
	perm, found := app.Permissions[method]
	if !found || perm == Ask {
		perm = x.requestHandler(app, request)

		if perm == AlwaysDeny || (perm == AlwaysAllow && x.CanStorePermission(method)) {
			app.Permissions[method] = perm
		}

		if perm.IsPositive() {
			x.logger.Info("Permission granted", "method", method, "permission", perm)
		} else {
			x.logger.Info("Permission rejected", "method", method, "permission", perm)
		}
	} else {
		x.logger.V(1).Info("Permission already granted for method", "method", method, "permission", perm)
	}

	return perm
}

// block until the session is closed and read all its messages
func (x *XSWD) readMessageFromSession(conn *Connection, app *ApplicationData) {
	defer x.removeApplicationOfSession(conn, app)

	for {
		// Remove application if it exceeds request rate limit
		if app.limiter != nil && !app.limiter.Allow() {
			x.logger.Error(fmt.Errorf("requests have exceeded rate limit"), "Rate limit exceeded", app.Name, "closing connection")
			if err := conn.Send(ResponseWithError(nil, jrpc2.Errorf(RateLimitExceeded, "Requests have exceeded rate limit, closing connection"))); err != nil {
				return
			}

			return
		}

		// block and read the message bytes from session
		_, buff, err := conn.Read()
		if err != nil {
			x.logger.V(2).Error(err, "Error while reading message from session")
			return
		}

		// app tried to send us a request while he was not authorized yet
		if !x.HasApplicationId(app.Id) {
			x.logger.Info("App is not authorized and requests us, closing connection")
			return
		}

		// unmarshal the request
		requests, err := jrpc2.ParseRequests(buff)
		if err != nil {
			x.logger.Error(err, "Error while parsing request")
			if err := conn.Send(ResponseWithError(nil, jrpc2.Errorf(code.ParseError, "Error while parsing request"))); err != nil {
				return
			}
			continue
		}

		request := requests[0]
		// We only support one request at a time for permission request
		if len(requests) != 1 {
			x.logger.V(2).Error(nil, "Invalid number of requests")
			if err := conn.Send(ResponseWithError(nil, jrpc2.Errorf(code.ParseError, "Batch requests are not supported"))); err != nil {
				return
			}
			continue
		}

		// Convert ParsedRequest to Request
		req := request.ToRequest()
		if req == nil {
			x.logger.Error(nil, "Invalid request")
			if err := conn.Send(ResponseWithError(nil, jrpc2.Errorf(code.ParseError, "Invalid request"))); err != nil {
				return
			}
			continue
		}

		x.requests <- messageRequest{app: app, request: req, conn: conn}
	}
}

// Handle a WebSocket connection
func (x *XSWD) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	globals.Logger.V(2).Info("New WebSocket connection", "addr", r.RemoteAddr)
	// Accept from any origin
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		x.logger.V(1).Error(err, "WebSocket upgrade error")
		return
	}
	defer conn.Close()

	// first message of the session should be its ApplicationData
	var app_data ApplicationData
	if err := conn.ReadJSON(&app_data); err != nil {
		x.logger.V(2).Error(err, "Error while reading app_data")
		conn.WriteJSON(AuthorizationResponse{
			Message:  "Invalid app data format",
			Accepted: false,
		})

		return
	}

	if x.HasApplicationId(app_data.Id) {
		x.logger.Info("App ID is already used", "ID", app_data.Name)
		conn.WriteJSON(AuthorizationResponse{
			Message:  "App ID is already used",
			Accepted: false,
		})

		return
	}

	connection := new(Connection)
	connection.conn = conn
	x.registers <- messageRegistration{conn: connection, request: r, app: &app_data}
	x.readMessageFromSession(connection, &app_data)
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}
