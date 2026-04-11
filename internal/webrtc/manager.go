package webrtc

import (
	"QunDev/GoRemoteDesktop_Server/internal/protocol"
	"fmt"
	"log"
	"net"
	"os/exec"

	"github.com/pion/webrtc/v4"
)

func CreateHostOffer(onCandidate protocol.OnICECandidateFunc) (*webrtc.PeerConnection, *webrtc.SessionDescription, error) {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, nil, fmt.Errorf("NewPeerConnection: %w", err)
	}

	videoTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{
			MimeType:    webrtc.MimeTypeH264,
			ClockRate:   90000,
			SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f",
		},
		"video",
		"screen-capture",
	)
	if err != nil {
		return nil, nil, fmt.Errorf("NewTrackLocalStaticRTP: %w", err)
	}

	if _, err = peerConnection.AddTrack(videoTrack); err != nil {
		return nil, nil, fmt.Errorf("AddTrack: %w", err)
	}

	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			log.Println("[ICE] Gathering complete")
			return
		}
		//log.Printf("[ICE] candidate: %s | type: %s | addr: %s:%d",
		//	c.ToJSON().Candidate,
		//	c.Typ,
		//	c.Address,
		//	c.Port,
		//)
		if onCandidate != nil {
			onCandidate(c)
		}
	})

	go startFFmpegPipeline(videoTrack)

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("CreateOffer: %w", err)
	}

	if err = peerConnection.SetLocalDescription(offer); err != nil {
		return nil, nil, fmt.Errorf("SetLocalDescription: %w", err)
	}

	finalSDP := peerConnection.LocalDescription()

	log.Println("[Host] Offer sent. Waiting for answer...")
	return peerConnection, finalSDP, nil
}

func startFFmpegPipeline(track *webrtc.TrackLocalStaticRTP) {
	ready := make(chan struct{})
	go forwardRTPToTrack("127.0.0.1:5004", track, ready)
	<-ready
	cmd := exec.Command("ffmpeg",
		"-f", "gdigrab",
		"-framerate", "30",
		"-i", "desktop",
		"-vcodec", "libx264",
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-profile:v", "baseline",
		"-level", "3.1",
		"-b:v", "2M",
		"-maxrate", "2M",
		"-bufsize", "4M",
		"-pix_fmt", "yuv420p",
		"-g", "30",
		"-f", "rtp",
		"-sdp_file", "stream.sdp",
		"rtp://127.0.0.1:5004",
	)

	if err := cmd.Start(); err != nil {
		log.Fatalf("[FFmpeg] Start failed: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		log.Printf("[FFmpeg] Exited: %v", err)
	}
}

func forwardRTPToTrack(addr string, track *webrtc.TrackLocalStaticRTP, ready chan struct{}) {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		log.Fatalf("ListenPacket: %v", err)
	}
	defer conn.Close()

	log.Printf("[RTP] Listening on %s", addr)
	close(ready)

	buf := make([]byte, 1500)
	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			log.Printf("[RTP] Read error: %v", err)
			return
		}
		if _, err = track.Write(buf[:n]); err != nil {
			log.Printf("[RTP] Write error: %v", err)
			return
		}
	}
}

func watchConnectionState(pc *webrtc.PeerConnection) {
	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Printf("[ICE] %s", state)
		if state == webrtc.ICEConnectionStateFailed {
			log.Println("[ICE] Failed — closing")
			pc.Close()
		}
	})

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Printf("[PC] %s", state)
		switch state {
		case webrtc.PeerConnectionStateConnected:
			log.Println("[PC] Streaming ✓")
		case webrtc.PeerConnectionStateFailed:
			log.Println("[PC] Failed — closing")
			pc.Close()
		}
	})
}
