package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

var defaultStunServers = []string{
	"stun:49.13.204.141:3478",
	"stun:176.58.93.154:3478",
	"stun:185.40.234.113:3478",
	"stun:68.183.90.120:3478",
	"stun:45.159.97.233:3478",
	"stun:172.105.166.103:3478",
	"stun:172.237.28.183:3478",
	"stun:208.72.155.133:3478",
	"stun:stun.l.google.com:19302",
}

func getStunServers() []string {
	if env := os.Getenv("STUN_SERVERS"); env != "" {
		return strings.Split(env, ",")
	}
	return defaultStunServers
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envIntOrDefault(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getFfmpegPath() string {
	return envOrDefault("FFMPEG_PATH", "ffmpeg")
}

type Peer struct {
	ID         string
	PC         *webrtc.PeerConnection
	VideoTrack *webrtc.TrackLocalStaticRTP
	AudioTrack *webrtc.TrackLocalStaticRTP
	Active     bool
	mu         sync.Mutex
}

type Sidecar struct {
	peers     map[string]*Peer
	peersLock sync.RWMutex

	videoPort int
	audioPort int
	videoConn *net.UDPConn
	audioConn *net.UDPConn

	ffmpeg     *exec.Cmd
	ffmpegLock sync.Mutex
	source     string
	running    bool

	videoTsFirst uint32
	videoTsGot   bool
	audioTsFirst uint32
	audioTsGot   bool
	tsLock       sync.Mutex
}

func NewSidecar() *Sidecar {
	return &Sidecar{
		peers: make(map[string]*Peer),
	}
}

func (s *Sidecar) resetTimestamps() {
	s.tsLock.Lock()
	s.videoTsGot = false
	s.audioTsGot = false
	s.tsLock.Unlock()
	log.Printf("[SYNC] Timestamp normalization reset")
}

func (s *Sidecar) StartRTP() error {
	var err error

	s.videoConn, err = net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		return fmt.Errorf("bind video UDP: %w", err)
	}
	s.videoPort = s.videoConn.LocalAddr().(*net.UDPAddr).Port

	s.audioConn, err = net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		return fmt.Errorf("bind audio UDP: %w", err)
	}
	s.audioPort = s.audioConn.LocalAddr().(*net.UDPAddr).Port

	log.Printf("[RTP] Video port: %d, Audio port: %d", s.videoPort, s.audioPort)
	s.running = true

	go s.readVideoRTP()
	go s.readAudioRTP()

	return nil
}

func (s *Sidecar) readVideoRTP() {
	buf := make([]byte, 1500)
	pkt := &rtp.Packet{}
	count := 0

	for s.running {
		n, err := s.videoConn.Read(buf)
		if err != nil {
			if s.running {
				log.Printf("[RTP] Video read error: %v", err)
			}
			return
		}

		if err := pkt.Unmarshal(buf[:n]); err != nil {
			continue
		}

		// Simply subtract the first timestamp to normalize to zero-based.
		// FFmpeg already guarantees A/V sync within a single invocation,
		// so no cross-stream wallclock correction is needed.
		s.tsLock.Lock()
		if !s.videoTsGot {
			s.videoTsFirst = pkt.Timestamp
			s.videoTsGot = true
			log.Printf("[VIDEO] First ts=%d", pkt.Timestamp)
		}
		first := s.videoTsFirst
		s.tsLock.Unlock()

		pkt.Timestamp = pkt.Timestamp - first
		count++

		if count <= 3 || count%600 == 0 {
			log.Printf("[VIDEO] #%d ts=%d (%.3fs) marker=%v",
				count, pkt.Timestamp, float64(pkt.Timestamp)/90000.0, pkt.Marker)
		}

		s.peersLock.RLock()
		for _, peer := range s.peers {
			if peer.Active && peer.VideoTrack != nil {
				peer.VideoTrack.WriteRTP(pkt)
			}
		}
		s.peersLock.RUnlock()
	}
}

func (s *Sidecar) readAudioRTP() {
	buf := make([]byte, 1500)
	pkt := &rtp.Packet{}
	count := 0

	for s.running {
		n, err := s.audioConn.Read(buf)
		if err != nil {
			if s.running {
				log.Printf("[RTP] Audio read error: %v", err)
			}
			return
		}

		if err := pkt.Unmarshal(buf[:n]); err != nil {
			continue
		}

		// Simply subtract the first timestamp to normalize to zero-based.
		s.tsLock.Lock()
		if !s.audioTsGot {
			s.audioTsFirst = pkt.Timestamp
			s.audioTsGot = true
			log.Printf("[AUDIO] First ts=%d", pkt.Timestamp)
		}
		first := s.audioTsFirst
		s.tsLock.Unlock()

		pkt.Timestamp = pkt.Timestamp - first
		count++

		if count <= 3 || count%1000 == 0 {
			log.Printf("[AUDIO] #%d ts=%d (%.3fs)",
				count, pkt.Timestamp, float64(pkt.Timestamp)/48000.0)
		}

		s.peersLock.RLock()
		for _, peer := range s.peers {
			if peer.Active && peer.AudioTrack != nil {
				peer.AudioTrack.WriteRTP(pkt)
			}
		}
		s.peersLock.RUnlock()
	}
}

func (s *Sidecar) CreatePeer(id string) (string, error) {
	iceServers := []webrtc.ICEServer{}
	for _, stun := range getStunServers() {
		iceServers = append(iceServers, webrtc.ICEServer{URLs: []string{stun}})
	}

	m := &webrtc.MediaEngine{}
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:    webrtc.MimeTypeVP8,
			ClockRate:   90000,
			SDPFmtpLine: "",
		},
		PayloadType: 96,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		return "", err
	}
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:  webrtc.MimeTypeOpus,
			ClockRate: 48000,
			Channels:  2,
		},
		PayloadType: 111,
	}, webrtc.RTPCodecTypeAudio); err != nil {
		return "", err
	}

	i := &interceptor.Registry{}
	intervalPliFactory, err := intervalpli.NewReceiverInterceptor()
	if err != nil {
		return "", err
	}
	i.Add(intervalPliFactory)
	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		return "", err
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithInterceptorRegistry(i))

	pc, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: iceServers,
	})
	if err != nil {
		return "", fmt.Errorf("create PeerConnection: %w", err)
	}

	videoTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8, ClockRate: 90000},
		"video", "ts6-stream",
	)
	if err != nil {
		pc.Close()
		return "", err
	}

	audioTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 2},
		"audio", "ts6-stream",
	)
	if err != nil {
		pc.Close()
		return "", err
	}

	if _, err = pc.AddTrack(videoTrack); err != nil {
		pc.Close()
		return "", err
	}
	if _, err = pc.AddTrack(audioTrack); err != nil {
		pc.Close()
		return "", err
	}

	peer := &Peer{
		ID:         id,
		PC:         pc,
		VideoTrack: videoTrack,
		AudioTrack: audioTrack,
		Active:     false,
	}

	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Printf("[Peer %s] ICE: %s", id, state.String())
		switch state {
		case webrtc.ICEConnectionStateConnected:
			peer.mu.Lock()
			peer.Active = true
			peer.mu.Unlock()
		case webrtc.ICEConnectionStateDisconnected, webrtc.ICEConnectionStateFailed, webrtc.ICEConnectionStateClosed:
			peer.mu.Lock()
			peer.Active = false
			peer.mu.Unlock()
		}
	})

	s.peersLock.Lock()
	if old, exists := s.peers[id]; exists {
		old.Active = false
		old.PC.Close()
	}
	s.peers[id] = peer
	s.peersLock.Unlock()

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		return "", fmt.Errorf("create offer: %w", err)
	}
	if err := pc.SetLocalDescription(offer); err != nil {
		return "", fmt.Errorf("set local desc: %w", err)
	}

	gatherComplete := webrtc.GatheringCompletePromise(pc)
	<-gatherComplete

	return pc.LocalDescription().SDP, nil
}

func (s *Sidecar) SetAnswer(id, sdp string) error {
	s.peersLock.RLock()
	peer, exists := s.peers[id]
	s.peersLock.RUnlock()
	if !exists {
		return fmt.Errorf("peer %s not found", id)
	}

	return peer.PC.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  sdp,
	})
}

func (s *Sidecar) AddICECandidate(id string, candidate string, sdpMid string, sdpMLineIndex uint16) error {
	s.peersLock.RLock()
	peer, exists := s.peers[id]
	s.peersLock.RUnlock()
	if !exists {
		return fmt.Errorf("peer %s not found", id)
	}

	return peer.PC.AddICECandidate(webrtc.ICECandidateInit{
		Candidate:     candidate,
		SDPMid:        &sdpMid,
		SDPMLineIndex: &sdpMLineIndex,
	})
}

func (s *Sidecar) ClosePeer(id string) {
	s.peersLock.Lock()
	if peer, exists := s.peers[id]; exists {
		peer.Active = false
		peer.PC.Close()
		delete(s.peers, id)
	}
	s.peersLock.Unlock()
}

func (s *Sidecar) StartFFmpeg(source string) {
	s.ffmpegLock.Lock()
	defer s.ffmpegLock.Unlock()

	s.StopFFmpegLocked()
	s.source = source

	args := []string{}

	if source != "" {
		if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
			args = append(args, "-reconnect", "1", "-reconnect_streamed", "1", "-reconnect_delay_max", "5")
		} else {
			args = append(args, "-stream_loop", "-1")
		}
		args = append(args, "-fflags", "+genpts+discardcorrupt", "-re", "-i", source)
	} else {
		w := envIntOrDefault("VIDEO_WIDTH", 1280)
		h := envIntOrDefault("VIDEO_HEIGHT", 720)
		args = append(args, "-re", "-f", "lavfi", "-i", fmt.Sprintf("color=c=black:s=%dx%d:r=1", w, h))
	}

	w := envIntOrDefault("VIDEO_WIDTH", 1280)
	h := envIntOrDefault("VIDEO_HEIGHT", 720)
	fps := envOrDefault("VIDEO_FRAMERATE", "30")
	vBitrate := envOrDefault("VIDEO_BITRATE", "1500k")

	if source != "" {
		vf := fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,format=yuv420p", w, h, w, h)
		args = append(args,
			"-map", "0:v:0",
			"-vf", vf,
			"-r", fps,
		)
	}
	args = append(args,
		"-pix_fmt", "yuv420p",
		"-c:v", "libvpx",
		"-cpu-used", "8",
		"-deadline", "realtime",
		"-lag-in-frames", "0",
		"-error-resilient", "1",
		"-b:v", vBitrate,
		"-maxrate", "2M",
		"-bufsize", "100k",
		"-keyint_min", "15",
		"-g", "15",
		"-auto-alt-ref", "0",
		"-payload_type", "96",
		"-ssrc", "11111111",
		"-f", "rtp",
		fmt.Sprintf("rtp://127.0.0.1:%d", s.videoPort),
	)

	if source != "" {
		aBitrate := envOrDefault("AUDIO_BITRATE", "128k")
		args = append(args,
			"-map", "0:a:0?",
			"-c:a", "libopus",
			"-b:a", aBitrate,
			"-ar", "48000",
			"-ac", "2",
			"-payload_type", "111",
			"-ssrc", "22222222",
			"-f", "rtp",
			fmt.Sprintf("rtp://127.0.0.1:%d", s.audioPort),
		)
	}

	s.resetTimestamps()

	log.Printf("[FFmpeg] Starting: source=%s video=:%d audio=:%d", source, s.videoPort, s.audioPort)

	cmd := exec.Command(getFfmpegPath(), args...)
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Printf("[FFmpeg] Start error: %v", err)
		return
	}
	s.ffmpeg = cmd

	go func() {
		err := cmd.Wait()
		log.Printf("[FFmpeg] Exited: %v", err)
	}()
}

func (s *Sidecar) StopFFmpegLocked() {
	if s.ffmpeg != nil && s.ffmpeg.Process != nil {
		s.ffmpeg.Process.Kill()
		s.ffmpeg = nil
	}
}

func (s *Sidecar) GetStats() map[string]interface{} {
	s.peersLock.RLock()
	defer s.peersLock.RUnlock()

	peers := map[string]interface{}{}
	for id, peer := range s.peers {
		peers[id] = map[string]interface{}{
			"active": peer.Active,
			"state":  peer.PC.ICEConnectionState().String(),
		}
	}

	return map[string]interface{}{
		"videoPort": s.videoPort,
		"audioPort": s.audioPort,
		"peerCount": len(s.peers),
		"peers":     peers,
		"source":    s.source,
	}
}

func (s *Sidecar) Stop() {
	s.running = false
	s.ffmpegLock.Lock()
	s.StopFFmpegLocked()
	s.ffmpegLock.Unlock()

	if s.videoConn != nil {
		s.videoConn.Close()
	}
	if s.audioConn != nil {
		s.audioConn.Close()
	}

	s.peersLock.Lock()
	for id, peer := range s.peers {
		peer.Active = false
		peer.PC.Close()
		delete(s.peers, id)
	}
	s.peersLock.Unlock()
}

func main() {
	port := 9800
	if p := os.Getenv("SIDECAR_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	sidecar := NewSidecar()
	if err := sidecar.StartRTP(); err != nil {
		log.Fatalf("Failed to start RTP: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /peer/create", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		log.Printf("[API] Creating peer: %s", req.ID)

		sdp, err := sidecar.CreatePeer(req.ID)
		if err != nil {
			log.Printf("[API] CreatePeer error: %v", err)
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"sdp": sdp})
	})

	mux.HandleFunc("POST /peer/answer", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID  string `json:"id"`
			SDP string `json:"sdp"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		log.Printf("[API] Setting answer for peer: %s (%d bytes)", req.ID, len(req.SDP))

		if err := sidecar.SetAnswer(req.ID, req.SDP); err != nil {
			log.Printf("[API] SetAnswer error: %v", err)
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /peer/ice", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID            string `json:"id"`
			Candidate     string `json:"candidate"`
			SDPMid        string `json:"sdpMid"`
			SDPMLineIndex uint16 `json:"sdpMLineIndex"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		if err := sidecar.AddICECandidate(req.ID, req.Candidate, req.SDPMid, req.SDPMLineIndex); err != nil {
			log.Printf("[API] AddICE error: %v", err)
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /peer/close", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		sidecar.ClosePeer(req.ID)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /source", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Source string `json:"source"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		log.Printf("[API] Setting source: %s", req.Source)
		sidecar.StartFFmpeg(req.Source)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /source/stop", func(w http.ResponseWriter, r *http.Request) {
		sidecar.ffmpegLock.Lock()
		sidecar.StopFFmpegLocked()
		sidecar.ffmpegLock.Unlock()
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /stats", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(sidecar.GetStats())
	})

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "ok",
			"videoPort": sidecar.videoPort,
			"audioPort": sidecar.audioPort,
		})
	})

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		sidecar.Stop()
		os.Exit(0)
	}()

	log.Printf("[Sidecar] HTTP API listening on :%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}
