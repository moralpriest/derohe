// Copyright 2017-2021 DERO Project. All rights reserved.
// Use of this source code in any form is governed by RESEARCH license.
// license can be found in the LICENSE file.
// GPG: 0F39 E425 8C65 3947 702A  8234 08B2 0360 A03A 9DE8
//
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY
// EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL
// THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
// PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
// STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF
// THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package p2p

import "time"
import "math/rand"

import "github.com/beevik/ntp"
import "github.com/go-logr/logr"

import "github.com/deroproject/derohe/globals"

// these servers automatically rotate every hour as per documentation
// we also rotate them randomly
// TODO support ipv6
var timeservers = []string{ // facebook/google do leap smearing, so they should not be mixed here
	"0.pool.ntp.org",
	"1.pool.ntp.org",
	"2.pool.ntp.org",
	"3.pool.ntp.org",
	"ntp1.hetzner.de",
	"ntp2.hetzner.de",
	"ntp3.hetzner.de",
	"time.cloudflare.com", // anycast
	"ntp.se",              // anycast
}

const clockDriftThreshold = time.Second

// clockTracker emits loud, once-per-transition warnings (default log level).
var clockTracker = &clockState{}

type clockState struct {
	driftWarned       bool
	unreachableWarned bool
}

// observe records one NTP attempt. ntpOK false means no server answered.
// Warnings print at default verbosity, once per state change.
func (s *clockState) observe(ntpOK bool, offset time.Duration, log logr.Logger) {
	if log.GetSink() == nil {
		return
	}
	if !ntpOK {
		if !s.unreachableWarned {
			s.unreachableWarned = true
			log.Error(nil, "Cannot reach NTP servers (UDP/123). Clock cannot be verified. Allow outbound NTP or install chrony.")
		}
		return
	}
	s.unreachableWarned = false
	drifted := offset > clockDriftThreshold || offset < -clockDriftThreshold
	if drifted {
		if !s.driftWarned {
			s.driftWarned = true
			log.Error(nil, "CLOCK DRIFT: system time is more than 1s off NTP. Chain sync and mining rewards may fail. Sync with chrony/NTP (e.g. timedatectl set-ntp true / chronyc tracking).", "offset", offset)
		}
		return
	}
	if s.driftWarned {
		s.driftWarned = false
		log.Info("Clock is back in sync with NTP.", "offset", offset)
	}
}

func applyNTPOffset(offset time.Duration) {
	if offset.Milliseconds() < -50 || offset.Milliseconds() > 50 {
		globals.ClockOffsetNTP = offset
	} else {
		globals.ClockOffsetNTP = 0
	}
	globals.TimeIsInSyncNTP = true
}

func queryOneNTP(server string) (time.Duration, error) {
	response, err := ntp.Query(server)
	if err != nil {
		return 0, err
	}
	if err := response.Validate(); err != nil {
		return 0, err
	}
	return response.ClockOffset, nil
}

// probeClockOnce runs a short synchronous NTP check before the background loop
// so a drifted clock is reported immediately at startup.
func probeClockOnce() {
	log := logger
	if log.GetSink() == nil {
		log = globals.Logger
	}
	for i := 0; i < 3 && i < len(timeservers); i++ {
		offset, err := queryOneNTP(timeservers[i])
		if err != nil {
			continue
		}
		applyNTPOffset(offset)
		clockTracker.observe(true, offset, log)
		return
	}
	clockTracker.observe(false, 0, log)
}

// continuously checks time for deviation if possible
func time_check_routine() {
	const offset_count = 128
	var offsets [offset_count]time.Duration
	var offset_index int

	random := rand.New(globals.NewCryptoRandSource())
	timeinsync := false
	for {
		server := timeservers[random.Int()%len(timeservers)]

		if offset, err := queryOneNTP(server); err != nil {
			clockTracker.observe(false, 0, logger)
		} else {
			offsets[offset_index] = offset
			offset_index = (offset_index + 1) % offset_count

			var avg_offset time.Duration
			var avg_count time.Duration
			for _, o := range offsets {
				if o != 0 {
					avg_offset += o
					avg_count++
				}
			}
			if avg_count > 0 {
				avg_offset = avg_offset / avg_count
			}
			applyNTPOffset(avg_offset)
			if offset > -clockDriftThreshold && offset < clockDriftThreshold {
				timeinsync = true
			} else {
				timeinsync = false
			}
			clockTracker.observe(true, offset, logger)
		}

		if !timeinsync {
			time.Sleep(5 * time.Second)
		} else {
			time.Sleep(time.Duration((random.Intn(60) + 60)) * time.Second)
		}
	}
}
