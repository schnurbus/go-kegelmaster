//go:build embed

package server

import (
	"embed"
	"io/fs"
	"path"
	"strings"

	"github.com/gofiber/fiber/v3"
)

//go:embed web/*
var embedFS embed.FS

// registerStatic registers static file serving and SPA fallback for the embedded frontend.
// API routes are registered first, so /api/* is handled by the API; only non-API GETs reach here.
func registerStatic(s *Server) {
	sub, _ := fs.Sub(embedFS, "web")

	s.app.Get("/*", func(c fiber.Ctx) error {
		p := strings.TrimPrefix(c.Path(), "/")
		if p == "" {
			p = "index.html"
		}
		f, err := sub.Open(p)
		if err != nil {
			// SPA fallback: serve index.html for client-side routing
			data, err := fs.ReadFile(sub, "index.html")
			if err != nil {
				return c.SendStatus(fiber.StatusNotFound)
			}
			c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
			return c.Send(data)
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil {
			data, _ := fs.ReadFile(sub, "index.html")
			c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
			return c.Send(data)
		}
		if info.IsDir() {
			p = strings.TrimSuffix(p, "/") + "/index.html"
			data, err := fs.ReadFile(sub, p)
			if err != nil {
				data, _ = fs.ReadFile(sub, "index.html")
				c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
				return c.Send(data)
			}
			c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
			return c.Send(data)
		}
		c.Set(fiber.HeaderContentType, contentTypeByExt(path.Ext(p)))
		return c.SendStream(f, int(info.Size()))
	})
}

func contentTypeByExt(ext string) string {
	switch ext {
	case ".html":
		return fiber.MIMETextHTMLCharsetUTF8
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".json":
		return fiber.MIMEApplicationJSONCharsetUTF8
	case ".ico":
		return "image/x-icon"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".woff2":
		return "font/woff2"
	case ".woff":
		return "font/woff"
	default:
		return "application/octet-stream"
	}
}
