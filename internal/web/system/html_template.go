package system

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>System Error Reference</title>
    <link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;500;600;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-primary: #030712;
            --bg-secondary: #111827;
            --glass-bg: rgba(17, 24, 39, 0.7);
            --glass-border: rgba(255, 255, 255, 0.08);
            --text-primary: #f3f4f6;
            --text-secondary: #9ca3af;
            --accent-primary: #3b82f6;
            --accent-glow: rgba(59, 130, 246, 0.5);
            
            --cat-data: #818cf8;
            --cat-validation: #fbbf24;
            --cat-security: #f87171;
            --cat-system: #38bdf8;
            --cat-business: #34d399;
        }

        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            font-family: 'Outfit', sans-serif;
            background-color: var(--bg-primary);
            background-image: 
                radial-gradient(circle at 15% 50%, rgba(59, 130, 246, 0.08), transparent 25%),
                radial-gradient(circle at 85% 30%, rgba(139, 92, 246, 0.08), transparent 25%);
            color: var(--text-primary);
            line-height: 1.6;
            min-height: 100vh;
            -webkit-font-smoothing: antialiased;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 4rem 2rem;
        }

        header {
            text-align: center;
            margin-bottom: 4rem;
            animation: fadeInDown 0.8s ease-out;
        }

        h1 {
            font-size: 3.5rem;
            font-weight: 700;
            letter-spacing: -0.02em;
            margin-bottom: 1rem;
            background: linear-gradient(135deg, #fff 0%, #94a3b8 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            text-shadow: 0 0 30px rgba(255,255,255,0.1);
        }

        .subtitle {
            color: var(--text-secondary);
            font-size: 1.25rem;
            max-width: 600px;
            margin: 0 auto;
            font-weight: 300;
        }

        /* Search Section */
        .controls {
            position: sticky;
            top: 2rem;
            z-index: 100;
            margin-bottom: 3rem;
            animation: fadeInUp 0.8s ease-out 0.2s backwards;
        }

        .search-container {
            background: var(--glass-bg);
            backdrop-filter: blur(20px);
            -webkit-backdrop-filter: blur(20px);
            padding: 1rem;
            border-radius: 20px;
            border: 1px solid var(--glass-border);
            box-shadow: 
                0 4px 6px -1px rgba(0, 0, 0, 0.1), 
                0 2px 4px -1px rgba(0, 0, 0, 0.06),
                0 0 0 1px rgba(255,255,255,0.05) inset;
            display: flex;
            align-items: center;
            gap: 1rem;
            max-width: 800px;
            margin: 0 auto;
            transition: transform 0.2s, box-shadow 0.2s;
        }

        .search-container:focus-within {
            transform: translateY(-2px);
            box-shadow: 
                0 20px 25px -5px rgba(0, 0, 0, 0.2), 
                0 10px 10px -5px rgba(0, 0, 0, 0.1),
                0 0 0 1px rgba(59, 130, 246, 0.5) inset;
        }

        .search-icon {
            color: var(--text-secondary);
            font-size: 1.2rem;
            padding-left: 1rem;
        }

        .search-input {
            width: 100%;
            background: transparent;
            border: none;
            color: var(--text-primary);
            font-size: 1.1rem;
            font-family: inherit;
            padding: 0.5rem;
        }

        .search-input:focus {
            outline: none;
        }

        .search-input::placeholder {
            color: rgba(156, 163, 175, 0.5);
        }

        /* Accordions */
        .accordions-wrapper {
            display: flex;
            flex-direction: column;
            gap: 1.5rem;
            animation: fadeInUp 0.8s ease-out 0.4s backwards;
        }

        .accordion-item {
            background: rgba(30, 41, 59, 0.4);
            border: 1px solid var(--glass-border);
            border-radius: 16px;
            overflow: hidden;
            transition: all 0.3s ease;
        }

        .accordion-item:hover {
            border-color: rgba(255,255,255,0.1);
            background: rgba(30, 41, 59, 0.6);
        }

        .accordion-header {
            width: 100%;
            padding: 1.5rem 2rem;
            background: none;
            border: none;
            display: flex;
            justify-content: space-between;
            align-items: center;
            cursor: pointer;
            color: var(--text-primary);
            transition: background 0.2s;
        }

        .category-title {
            display: flex;
            align-items: center;
            gap: 1rem;
            font-size: 1.25rem;
            font-weight: 600;
        }

        .category-dot {
            width: 10px;
            height: 10px;
            border-radius: 50%;
            position: relative;
        }
        
        .category-dot::after {
            content: '';
            position: absolute;
            top: -4px;
            left: -4px;
            right: -4px;
            bottom: -4px;
            border-radius: 50%;
            opacity: 0.3;
            background: inherit;
            filter: blur(4px);
        }

        .dot-Data { background: var(--cat-data); }
        .dot-Validation { background: var(--cat-validation); }
        .dot-Security { background: var(--cat-security); }
        .dot-System { background: var(--cat-system); }
        .dot-Business { background: var(--cat-business); }

        .category-info {
            display: flex;
            align-items: center;
            gap: 1.5rem;
        }

        .category-count {
            font-size: 0.9rem;
            color: var(--text-secondary);
            background: rgba(255,255,255,0.05);
            padding: 0.25rem 0.75rem;
            border-radius: 20px;
        }

        .accordion-icon {
            color: var(--text-secondary);
            transition: transform 0.3s ease;
        }

        .accordion-item.open .accordion-icon {
            transform: rotate(180deg);
        }

        .accordion-content {
            height: 0;
            overflow: hidden;
            transition: height 0.4s cubic-bezier(0.4, 0, 0.2, 1);
        }

        .accordion-inner {
            padding: 0 2rem 2rem 2rem;
        }

        /* Error Grid */
        .error-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
            gap: 1rem;
            padding-top: 1rem;
        }

        .error-card {
            background: rgba(255, 255, 255, 0.03);
            border: 1px solid rgba(255, 255, 255, 0.05);
            border-radius: 12px;
            padding: 1.5rem;
            cursor: pointer;
            transition: all 0.2s ease;
            position: relative;
            overflow: hidden;
        }

        .error-card::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            width: 3px;
            height: 100%;
            background: var(--text-secondary);
            opacity: 0.5;
            transition: all 0.2s;
        }

        .cat-Data .error-card::before { background: var(--cat-data); }
        .cat-Validation .error-card::before { background: var(--cat-validation); }
        .cat-Security .error-card::before { background: var(--cat-security); }
        .cat-System .error-card::before { background: var(--cat-system); }
        .cat-Business .error-card::before { background: var(--cat-business); }

        .error-card:hover {
            transform: translateY(-4px);
            background: rgba(255, 255, 255, 0.06);
            border-color: rgba(255, 255, 255, 0.1);
            box-shadow: 0 10px 20px -5px rgba(0,0,0,0.3);
        }

        .error-card:hover::before {
            opacity: 1;
            width: 4px;
        }

        .card-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 1rem;
        }

        .code-pill {
            font-family: 'Monaco', 'Consolas', monospace;
            font-size: 0.9rem;
            font-weight: 700;
            letter-spacing: -0.5px;
            color: var(--text-primary);
        }

        .http-badge {
            font-size: 0.75rem;
            font-weight: 600;
            padding: 0.25rem 0.5rem;
            border-radius: 6px;
            background: rgba(255,255,255,0.1);
            color: var(--text-secondary);
        }
        
        .status-5xx { background: rgba(248, 113, 113, 0.15); color: #fca5a5; }
        .status-4xx { background: rgba(251, 191, 36, 0.15); color: #fcd34d; }

        .error-message {
            color: #e2e8f0;
            font-size: 0.95rem;
            margin-bottom: 1rem;
            line-height: 1.5;
        }

        .error-meta {
            display: flex;
            justify-content: space-between;
            align-items: center;
            font-size: 0.8rem;
            color: var(--text-secondary);
            border-top: 1px solid rgba(255, 255, 255, 0.05);
            padding-top: 0.75rem;
        }

        /* Toast */
        .toast {
            position: fixed;
            bottom: 2rem;
            right: 2rem;
            background: var(--bg-secondary);
            border: 1px solid var(--accent-primary);
            color: var(--text-primary);
            padding: 1rem 2rem;
            border-radius: 8px;
            box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
            transform: translateY(150%);
            transition: transform 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
            z-index: 1000;
            display: flex;
            align-items: center;
            gap: 0.75rem;
        }
        
        .toast.show {
            transform: translateY(0);
        }

        /* Animations */
        @keyframes fadeInDown {
            from { opacity: 0; transform: translateY(-20px); }
            to { opacity: 1; transform: translateY(0); }
        }

        @keyframes fadeInUp {
            from { opacity: 0; transform: translateY(20px); }
            to { opacity: 1; transform: translateY(0); }
        }

        .empty-state {
            text-align: center;
            padding: 4rem;
            color: var(--text-secondary);
            display: none;
        }

        /* Mobile */
        @media (max-width: 768px) {
            h1 { font-size: 2.5rem; }
            .container { padding: 2rem 1rem; }
            .accordion-header { padding: 1.25rem; }
            .accordion-inner { padding: 0 1.25rem 1.25rem 1.25rem; }
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>System Error Reference</h1>
            <div class="subtitle">Comprehensive catalog of application status codes and messages</div>
        </header>

        <div class="controls">
            <div class="search-container">
                <span class="search-icon">⌕</span>
                <input type="text" id="searchInput" class="search-input" placeholder="Search by code, message, or ID..." autofocus>
            </div>
        </div>
        
        <div id="noResults" class="empty-state">
            <div style="font-size: 3rem; margin-bottom: 1rem;">👻</div>
            <p>No errors matching your criteria.</p>
        </div>

        <div class="accordions-wrapper">
            {{range .Categories}}
            <div class="accordion-item open cat-{{.Name}} search-group" data-category="{{.Name}}">
                <button class="accordion-header" onclick="toggleAccordion(this)">
                    <div class="category-title">
                        <span class="category-dot dot-{{.Name}}"></span>
                        {{.Name}}
                    </div>
                    <div class="category-info">
                        <span class="category-count">{{len .Errors}} codes</span>
                        <span class="accordion-icon">▼</span>
                    </div>
                </button>
                <div class="accordion-content" style="height: auto;"> 
                    <div class="accordion-inner">
                        <div class="error-grid">
                            {{range .Errors}}
                            <div class="error-card search-item" 
                                 data-code="{{.Code}}" 
                                 data-message="{{.Message}}" 
                                 data-id="{{.NumericCode}}"
                                 onclick="copyCode('{{.Code}}')">
                                <div class="card-header">
                                    <span class="code-pill">{{.Code}}</span>
                                    <span class="http-badge {{if ge .HTTPStatus 500}}status-5xx{{else}}status-4xx{{end}}">
                                        HTTP {{.HTTPStatus}}
                                    </span>
                                </div>
                                <div class="error-message">{{.Message}}</div>
                                <div class="error-meta">
                                    <span>ID: {{.NumericCode}}</span>
                                    <span>{{.Layer}}</span>
                                </div>
                            </div>
                            {{end}}
                        </div>
                    </div>
                </div>
            </div>
            {{end}}
        </div>
    </div>

    <!-- Toast Notification -->
    <div id="toast" class="toast">
        <span style="color: var(--cat-business);">✓</span>
        <span id="toastMsg">Copied to clipboard</span>
    </div>

    <script>
        function toggleAccordion(button) {
            const item = button.parentElement;
            const content = button.nextElementSibling;
            
            if (item.classList.contains('open')) {
                content.style.height = content.scrollHeight + 'px';
                content.offsetHeight; // Force repaint
                content.style.height = '0';
                item.classList.remove('open');
            } else {
                item.classList.add('open');
                content.style.height = content.scrollHeight + 'px';
                setTimeout(() => {
                    if(item.classList.contains('open')) {
                        content.style.height = 'auto';
                    }
                }, 400);
            }
        }

        function copyCode(code) {
            navigator.clipboard.writeText(code).then(() => {
                showToast(code);
            });
        }

        function showToast(code) {
            const toast = document.getElementById('toast');
            const msg = document.getElementById('toastMsg');
            msg.textContent = code + ' copied!';
            toast.classList.add('show');
            setTimeout(() => {
                toast.classList.remove('show');
            }, 2000);
        }

        const searchInput = document.getElementById('searchInput');
        const groups = document.querySelectorAll('.search-group');
        const noResults = document.getElementById('noResults');

        searchInput.addEventListener('input', (e) => {
            const term = e.target.value.toLowerCase();
            let totalVisible = 0;

            groups.forEach(group => {
                const items = group.querySelectorAll('.search-item');
                let foundInGroup = 0;

                items.forEach(item => {
                    const code = item.dataset.code.toLowerCase();
                    const msg = item.dataset.message.toLowerCase();
                    const id = item.dataset.id.toString();

                    if (code.includes(term) || msg.includes(term) || id.includes(term)) {
                        item.style.display = 'block';
                        foundInGroup++;
                    } else {
                        item.style.display = 'none';
                    }
                });

                if (foundInGroup > 0) {
                    group.style.display = 'block';
                    if (term.length > 0 && !group.classList.contains('open')) {
                         group.classList.add('open');
                         const content = group.querySelector('.accordion-content');
                         content.style.height = 'auto';
                    }
                } else {
                    group.style.display = 'none';
                }
                
                totalVisible += foundInGroup;
            });

            noResults.style.display = totalVisible === 0 ? 'block' : 'none';
        });
    </script>
</body>
</html>
`
