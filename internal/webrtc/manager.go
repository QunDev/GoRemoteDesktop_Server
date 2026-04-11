package webrtc

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"sync"

	"github.com/pion/webrtc/v4"
)

func CreateHostOffer() (*webrtc.PeerConnection, *webrtc.SessionDescription, error) {
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
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264},
		"video",
		"screen-capture",
	)
	if err != nil {
		return nil, nil, fmt.Errorf("NewTrackLocalStaticRTP: %w", err)
	}

	if _, err = peerConnection.AddTrack(videoTrack); err != nil {
		return nil, nil, fmt.Errorf("AddTrack: %w", err)
	}

	go startFFmpegPipeline(videoTrack)

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		log.Printf("[Host] Connection state: %s", s)
		if s == webrtc.PeerConnectionStateFailed {
			log.Println("[Host] Connection failed, closing...")
			_ = peerConnection.Close()
		}
	})

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("CreateOffer: %w", err)
	}

	if err = peerConnection.SetLocalDescription(offer); err != nil {
		return nil, nil, fmt.Errorf("SetLocalDescription: %w", err)
	}

	<-waitICEGathering(peerConnection)

	finalSDP := peerConnection.LocalDescription()

	log.Println("[Host] Offer sent. Waiting for answer...")
	return peerConnection, finalSDP, nil
}

func startFFmpegPipeline(track *webrtc.TrackLocalStaticRTP) {
	cmd := exec.Command("ffmpeg",
		"-f", "gdigrab", // input driver
		"-r", "30", // framerate
		"-s", "1920x1080", // độ phân giải
		"-i", "desktop", // display
		"-vcodec", "libx264",
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-b:v", "2M", // bitrate
		"-maxrate", "2M",
		"-bufsize", "4M",
		"-pix_fmt", "yuv420p",
		"-f", "rtp",
		"rtp://127.0.0.1:5004",
	)

	if err := cmd.Start(); err != nil {
		log.Fatalf("[FFmpeg] Start failed: %v", err)
	}

	go forwardRTPToTrack("127.0.0.1:5004", track)

	if err := cmd.Wait(); err != nil {
		log.Printf("[FFmpeg] Exited: %v", err)
	}
}

func forwardRTPToTrack(addr string, track *webrtc.TrackLocalStaticRTP) {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		log.Fatalf("ListenPacket: %v", err)
	}
	defer conn.Close()

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

func waitICEGathering(pc *webrtc.PeerConnection) <-chan struct{} {
	done := make(chan struct{})
	var once sync.Once

	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			once.Do(func() { close(done) })
		}
	})

	return done
}
