package restapi

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const reactAdminURL = "http://localhost:3000"

const adminRedirectHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Admin Panel</title>
    <meta http-equiv="refresh" content="2; url={{.URL}}">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "SF Pro Display", "Inter", sans-serif;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            background: linear-gradient(135deg, #E8F0FF 0%, #EEF2FF 30%, #EBF4FF 70%);
        }
        .card {
            background: rgba(255,255,255,0.72);
            backdrop-filter: blur(40px) saturate(180%);
            -webkit-backdrop-filter: blur(40px) saturate(180%);
            border: 1px solid rgba(255,255,255,0.6);
            border-radius: 24px;
            padding: 3rem 3.5rem;
            text-align: center;
            box-shadow: 0 8px 32px rgba(0,0,0,0.06);
            max-width: 440px;
        }
        .icon { font-size: 3rem; margin-bottom: 1.5rem; }
        h1 { font-size: 1.5rem; font-weight: 700; color: #1C1C1E; margin-bottom: 0.5rem; }
        p { color: #6C6C70; margin-bottom: 1.5rem; font-size: 0.95rem; }
        a {
            display: inline-block;
            background: linear-gradient(135deg, #007AFF, #0055CC);
            color: #fff;
            padding: 0.75rem 2rem;
            border-radius: 10px;
            text-decoration: none;
            font-weight: 600;
            font-size: 0.95rem;
        }
        .url { margin-top: 1rem; font-size: 0.8rem; color: #8E8E93; font-family: monospace; }
    </style>
</head>
<body>
    <div class="card">
        <div class="icon">⚙️</div>
        <h1>Admin Panel</h1>
        <p>Redirecting to the React admin panel...</p>
        <a href="{{.URL}}">Open Admin Panel →</a>
        <div class="url">{{.URL}}</div>
    </div>
</body>
</html>`

// setupAdminRedirect serves a redirect page at /admin pointing to the React SPA.
func setupAdminRedirect(handler *gin.Engine) {
	tmpl, _ := template.New("admin-redirect").Parse(adminRedirectHTML)

	redirectHandler := func(c *gin.Context) {
		data := struct{ URL string }{URL: reactAdminURL}
		var buf strings.Builder
		_ = tmpl.Execute(&buf, data)
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, buf.String())
	}

	handler.GET("/admin", redirectHandler)
	handler.GET("/admin/*path", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, reactAdminURL)
	})
}
