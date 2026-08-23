package p2p

import "bytes"
import "strings"
import "testing"
import "time"

import "github.com/go-logr/logr"
import "github.com/go-logr/zapr"
import "go.uber.org/zap"
import "go.uber.org/zap/zapcore"

func testLogger() (logr.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	enc := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(enc, zapcore.AddSync(buf), zapcore.InfoLevel)
	return zapr.NewLogger(zap.New(core)), buf
}

func TestClockState_WarnsOnceOnDrift(t *testing.T) {
	log, buf := testLogger()
	s := &clockState{}
	s.observe(true, 5*time.Second, log)
	s.observe(true, 6*time.Second, log)
	out := buf.String()
	if strings.Count(out, "CLOCK DRIFT") != 1 {
		t.Fatalf("expected exactly one CLOCK DRIFT warning, got:\n%s", out)
	}
}

func TestClockState_RecoveryOnce(t *testing.T) {
	log, buf := testLogger()
	s := &clockState{}
	s.observe(true, 5*time.Second, log)
	s.observe(true, 10*time.Millisecond, log)
	s.observe(true, 20*time.Millisecond, log)
	out := buf.String()
	if strings.Count(out, "CLOCK DRIFT") != 1 {
		t.Fatalf("expected one CLOCK DRIFT, got:\n%s", out)
	}
	if strings.Count(out, "back in sync") != 1 {
		t.Fatalf("expected one recovery message, got:\n%s", out)
	}
}

func TestClockState_UnreachableOnce(t *testing.T) {
	log, buf := testLogger()
	s := &clockState{}
	s.observe(false, 0, log)
	s.observe(false, 0, log)
	out := buf.String()
	if strings.Count(out, "Cannot reach NTP") != 1 {
		t.Fatalf("expected one unreachable warning, got:\n%s", out)
	}
}

func TestClockState_UnreachableClearsOnSuccess(t *testing.T) {
	log, buf := testLogger()
	s := &clockState{}
	s.observe(false, 0, log)
	s.observe(true, 5*time.Millisecond, log)
	s.observe(false, 0, log)
	out := buf.String()
	if strings.Count(out, "Cannot reach NTP") != 2 {
		t.Fatalf("expected unreachable to re-fire after a successful query, got:\n%s", out)
	}
}

func TestClockState_SmallOffsetSilent(t *testing.T) {
	log, buf := testLogger()
	s := &clockState{}
	s.observe(true, 200*time.Millisecond, log)
	if buf.Len() != 0 {
		t.Fatalf("small offset should be silent, got:\n%s", buf.String())
	}
}
