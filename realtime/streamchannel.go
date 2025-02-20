package realtime

import "sync"

type StreamChannel struct {
	Channel   chan []byte
	clients uint32
  connectionMu sync.Mutex
}

func NewStreamChannel() *StreamChannel {
	return &StreamChannel{
		clients: 0,
		Channel:   make(chan []byte),
	}
}

func (s *StreamChannel) Add() {
  s.connectionMu.Lock()
  s.clients++
  s.connectionMu.Unlock()
}

func (s *StreamChannel) Remove() {
  s.connectionMu.Lock()
  s.clients--
  s.connectionMu.Unlock()
}

func (s *StreamChannel) Connected() bool {
  s.connectionMu.Lock()
  defer s.connectionMu.Unlock()
  return s.clients > 0
}

func (s *StreamChannel) Send(data []byte) {
  s.connectionMu.Lock()
  clients := s.clients
  s.connectionMu.Unlock()

  // The assumption here is that the robot always only has one client.
  // We need to send these messages multiple times so that we can close
  // out all the other connections waiting on messages. Behaviour might
  // still be buggy.
  for range clients {
    s.Channel <- data
  }
}
