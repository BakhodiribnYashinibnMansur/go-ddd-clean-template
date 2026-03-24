package restapi

import (
	"bytes"
	"html/template"
	"net/http"

	"gct/config"
	"gct/internal/shared/infrastructure/httpx"

	"github.com/gin-gonic/gin"
)

// rootHTML defines the visual layout for the API landing page using Material Design aesthetics.
const rootHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Clean Template API</title>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;600;800&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg: #030712;
            --glass: rgba(17, 24, 39, 0.7);
            --primary: #38bdf8;
            --secondary: #818cf8;
            --accent: #f472b6;
            --white: #ffffff;
            --text: #f8fafc;
            --text-dim: #94a3b8;
            --success: #10b981;
        }
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: 'Inter', sans-serif;
            background-color: var(--bg);
            background-image: 
                radial-gradient(circle at 10% 20%, rgba(56, 189, 248, 0.08) 0%, transparent 40%),
                radial-gradient(circle at 90% 80%, rgba(129, 140, 248, 0.08) 0%, transparent 40%),
                radial-gradient(circle at 50% 50%, rgba(244, 114, 182, 0.03) 0%, transparent 50%);
            color: var(--text);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 1rem;
            overflow-x: hidden;
        }

        /* Animated Background Elements */
        .bg-shapes {
            position: fixed;
            top: 0; left: 0; width: 100%; height: 100%;
            pointer-events: none;
            z-index: -1;
            overflow: hidden;
        }
        .shape {
            position: absolute;
            background: linear-gradient(135deg, var(--primary), var(--secondary));
            border-radius: 50%;
            filter: blur(80px);
            opacity: 0.1;
            animation: float 20s infinite alternate cubic-bezier(0.45, 0, 0.55, 1);
        }
        .shape-1 { width: 400px; height: 400px; top: -100px; left: -100px; }
        .shape-2 { width: 300px; height: 300px; bottom: -50px; right: -50px; animation-delay: -5s; }
        
        @keyframes float {
            0% { transform: translate(0, 0) rotate(0deg) scale(1); }
            100% { transform: translate(50px, 100px) rotate(90deg) scale(1.1); }
        }

        .container {
            width: 100%;
            max-width: 900px;
            position: relative;
        }
        .glass-card {
            background: var(--glass);
            backdrop-filter: blur(24px);
            -webkit-backdrop-filter: blur(24px);
            border: 1px solid rgba(255, 255, 255, 0.08);
            border-radius: 40px;
            padding: 4rem 3rem;
            box-shadow: 
                0 25px 50px -12px rgba(0, 0, 0, 0.7),
                inset 0 1px 1px rgba(255, 255, 255, 0.05);
            text-align: center;
            animation: cardAppear 0.8s cubic-bezier(0.16, 1, 0.3, 1);
            position: relative;
            overflow: hidden;
        }

        @keyframes cardAppear {
            from { opacity: 0; transform: translateY(40px) scale(0.98); }
            to { opacity: 1; transform: translateY(0) scale(1); }
        }

        /* Header Animation */
        .header-content {
            animation: fadeInUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) both;
            animation-delay: 0.2s;
        }

        @keyframes fadeInUp {
            from { opacity: 0; transform: translateY(20px); }
            to { opacity: 1; transform: translateY(0); }
        }

        .badge {
            display: inline-flex;
            align-items: center;
            padding: 0.6rem 1.4rem;
            border-radius: 100px;
            background: rgba(16, 185, 129, 0.08);
            color: var(--success);
            font-size: 0.75rem;
            font-weight: 700;
            letter-spacing: 0.05em;
            margin-bottom: 2.5rem;
            border: 1px solid rgba(16, 185, 129, 0.15);
            text-transform: uppercase;
            transition: all 0.3s;
        }
        .badge:hover {
            background: rgba(16, 185, 129, 0.15);
            border-color: rgba(16, 185, 129, 0.3);
            transform: scale(1.05);
        }

        .status-dot {
            width: 8px; height: 8px;
            background: var(--success);
            border-radius: 50%;
            margin-right: 12px;
            box-shadow: 0 0 12px var(--success);
            animation: pulse 2s infinite;
        }
        @keyframes pulse {
            0% { transform: scale(1); opacity: 1; box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7); }
            70% { transform: scale(1.2); opacity: 0.8; box-shadow: 0 0 0 10px rgba(16, 185, 129, 0); }
            100% { transform: scale(1); opacity: 1; box-shadow: 0 0 0 0 rgba(16, 185, 129, 0); }
        }

        h1 {
            font-size: clamp(2.5rem, 8vw, 4.5rem);
            font-weight: 800;
            margin-bottom: 1.5rem;
            background: linear-gradient(135deg, #fff 0%, #94a3b8 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            letter-spacing: -0.05em;
        }
        .description {
            color: var(--text-dim);
            font-size: 1.2rem;
            margin-bottom: 3.5rem;
            line-height: 1.6;
            max-width: 650px;
            margin-left: auto;
            margin-right: auto;
        }

        /* Grid & Cards */
        .grid {
            display: grid;
            grid-template-columns: repeat(3, 1fr);
            gap: 1.5rem;
            margin-bottom: 3.5rem;
        }
        .card-link {
            display: flex;
            flex-direction: column;
            align-items: center;
            padding: 2.5rem 1.5rem;
            border-radius: 30px;
            text-decoration: none;
            transition: all 0.5s cubic-bezier(0.16, 1, 0.3, 1);
            background: rgba(255, 255, 255, 0.02);
            border: 1px solid rgba(255, 255, 255, 0.05);
            position: relative;
            overflow: hidden;
            opacity: 0;
            animation: fadeInUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) both;
        }
        .grid .card-link:nth-child(1) { animation-delay: 0.4s; }
        .grid .card-link:nth-child(2) { animation-delay: 0.5s; }
        .grid .card-link:nth-child(3) { animation-delay: 0.6s; }

        .card-link::before {
            content: '';
            position: absolute;
            top: 0; left: -100%;
            width: 100%; height: 100%;
            background: linear-gradient(to right, transparent, rgba(255,255,255,0.05), transparent);
            transition: 0.6s;
        }
        .card-link:hover::before { left: 100%; }

        .card-link:hover {
            transform: translateY(-12px) scale(1.02);
            background: rgba(255, 255, 255, 0.06);
            border-color: rgba(56, 189, 248, 0.3);
            box-shadow: 
                0 30px 60px -15px rgba(0,0,0,0.5),
                0 0 20px rgba(56, 189, 248, 0.1);
        }
        .card-link:active {
            transform: translateY(-4px) scale(0.96);
            transition: all 0.1s;
        }

        .icon-box {
            width: 64px; height: 64px;
            background: rgba(56, 189, 248, 0.06);
            border-radius: 20px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 2rem;
            margin-bottom: 1.5rem;
            transition: all 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275);
            border: 1px solid rgba(56, 189, 248, 0.1);
        }
        .card-link:hover .icon-box {
            transform: scale(1.1) rotate(8deg);
            background: rgba(56, 189, 248, 0.15);
            border-color: var(--primary);
            box-shadow: 0 0 15px rgba(56, 189, 248, 0.3);
        }

        .card-title {
            color: var(--white);
            font-weight: 700;
            font-size: 1.25rem;
            margin-bottom: 0.6rem;
            transition: color 0.3s;
        }
        .card-link:hover .card-title { color: var(--primary); }
        .card-desc {
            color: var(--text-dim);
            font-size: 0.9rem;
            line-height: 1.4;
        }

        .card-link.disabled {
            opacity: 0.3;
            cursor: not-allowed;
            pointer-events: none;
            filter: grayscale(1);
        }

        .footer-info {
            display: flex;
            justify-content: center;
            gap: 2.5rem;
            font-size: 0.875rem;
            color: var(--text-dim);
            padding-top: 2.5rem;
            border-top: 1px solid rgba(255, 255, 255, 0.05);
            animation: fadeInUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) both;
            animation-delay: 0.8s;
        }
        .info-item strong {
            color: var(--primary);
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }

        @media (max-width: 768px) {
            .grid { grid-template-columns: 1fr; }
            .glass-card { padding: 3rem 1.5rem; }
            h1 { font-size: 3rem; }
        }
    </style>
</head>
<body>
    <div class="bg-shapes">
        <div class="shape shape-1"></div>
        <div class="shape shape-2"></div>
    </div>
    <div class="container">
        <div class="glass-card">
            <div class="header-content">
                <div class="badge">
                    <span class="status-dot"></span>
                    API_CORE_ON_LINE
                </div>
                <h1>Go Clean Template</h1>
                <p class="description">
                    The pinnacle of Go architecture. Precision-engineered for scale, speed, and absolute reliability.
                </p>
            </div>
            
            <div class="grid">
                <a href="{{.SwaggerURL}}" target="_blank" class="card-link {{if not .SwaggerEnabled}}disabled{{end}}">
                    <div class="icon-box">📘</div>
                    <span class="card-title">Explorer</span>
                    <span class="card-desc">Interactive OpenAPI Documentation</span>
                </a>
                <a href="{{.ProtoURL}}" target="_blank" class="card-link {{if not .ProtoEnabled}}disabled{{end}}">
                    <div class="icon-box">📦</div>
                    <span class="card-title">Protobuf</span>
                    <span class="card-desc">Type-safe gRPC Definitions</span>
                </a>
                <a href="{{.AdminURL}}" target="_blank" class="card-link {{if not .AdminEnabled}}disabled{{end}}">
                    <div class="icon-box">⚙️</div>
                    <span class="card-title">Console</span>
                    <span class="card-desc">System Operations Control</span>
                </a>
            </div>

            <div class="footer-info">
                <div class="info-item">ENV: <strong>{{if .IsProduction}}PROD{{else}}DEV{{end}}</strong></div>
                <div class="info-item">VER: <strong>1.0.0</strong></div>
                <div class="info-item">REGION: <strong>GLOBAL</strong></div>
            </div>
        </div>
    </div>
</body>
</html>`

// setupRoot serves the visual API landing page with dynamic links to documentation.
func setupRoot(handler *gin.Engine, cfg *config.Config) {
	handler.GET("/", func(c *gin.Context) {
		scheme := "http"
		if c.Request.TLS != nil || httpx.GetForwardedProto(c) == "https" {
			scheme = "https"
		}

		data := struct {
			SwaggerURL, ProtoURL, AdminURL                           string
			SwaggerEnabled, ProtoEnabled, AdminEnabled, IsProduction bool
		}{
			SwaggerURL:     scheme + "://" + c.Request.Host + swaggerPath,
			ProtoURL:       scheme + "://" + c.Request.Host + protoPath,
			AdminURL:       scheme + "://" + c.Request.Host + adminPath,
			SwaggerEnabled: cfg.Swagger.Enabled || cfg.App.IsDev(),
			ProtoEnabled:   cfg.Proto.Enabled || cfg.App.IsDev(),
			AdminEnabled:   cfg.Admin.Enabled || cfg.App.IsDev(),
			IsProduction:   cfg.App.IsProd(),
		}

		tmpl, _ := template.New("root").Parse(rootHTML)
		var buf bytes.Buffer
		_ = tmpl.Execute(&buf, data)
		c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
	})
}
