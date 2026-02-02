//go:build !embed

package server

// registerStatic is a no-op when the embed build tag is not set.
func registerStatic(s *Server) {}
