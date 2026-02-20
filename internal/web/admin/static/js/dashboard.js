document.addEventListener('DOMContentLoaded', () => {
    initMainAreaChart();
    fetchDashboardStats();
    animateGauges();
    animateCounters();
});

// ─── Main Area Chart ─────────────────────────────────────────────────────────
function initMainAreaChart() {
    const canvas = document.getElementById('mainAreaChart');
    if (!canvas) return;

    const ctx = canvas.getContext('2d');

    // Gradient fills
    const gradientPrimary = ctx.createLinearGradient(0, 0, 0, 260);
    gradientPrimary.addColorStop(0, 'rgba(99, 102, 241, 0.25)');
    gradientPrimary.addColorStop(0.5, 'rgba(99, 102, 241, 0.08)');
    gradientPrimary.addColorStop(1, 'rgba(99, 102, 241, 0)');

    const gradientSecondary = ctx.createLinearGradient(0, 0, 0, 260);
    gradientSecondary.addColorStop(0, 'rgba(168, 85, 247, 0.2)');
    gradientSecondary.addColorStop(0.5, 'rgba(168, 85, 247, 0.05)');
    gradientSecondary.addColorStop(1, 'rgba(168, 85, 247, 0)');

    new Chart(ctx, {
        type: 'line',
        data: {
            labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
            datasets: [
                {
                    label: 'Requests',
                    data: [3200, 2800, 4100, 3800, 4200, 5100, 4700],
                    borderColor: '#818cf8',
                    backgroundColor: gradientPrimary,
                    borderWidth: 2,
                    fill: true,
                    tension: 0.4,
                    pointBackgroundColor: '#818cf8',
                    pointBorderColor: '#0c1021',
                    pointBorderWidth: 2,
                    pointRadius: 0,
                    pointHoverRadius: 5,
                    pointHoverBorderWidth: 2,
                    pointHoverBorderColor: '#fff'
                },
                {
                    label: 'Errors',
                    data: [120, 95, 180, 140, 160, 110, 90],
                    borderColor: '#a855f7',
                    backgroundColor: gradientSecondary,
                    borderWidth: 2,
                    fill: true,
                    tension: 0.4,
                    pointRadius: 0,
                    pointHoverRadius: 5,
                    pointBorderWidth: 2,
                    pointBackgroundColor: '#a855f7',
                    pointHoverBorderColor: '#fff'
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                mode: 'index',
                intersect: false
            },
            plugins: {
                legend: {
                    display: true,
                    position: 'top',
                    align: 'end',
                    labels: {
                        color: '#94a3b8',
                        font: { size: 11, family: 'Inter' },
                        boxWidth: 12,
                        boxHeight: 2,
                        padding: 16,
                        usePointStyle: false
                    }
                },
                tooltip: {
                    backgroundColor: 'rgba(12, 16, 33, 0.95)',
                    borderColor: 'rgba(99, 102, 241, 0.3)',
                    borderWidth: 1,
                    titleColor: '#f1f5f9',
                    bodyColor: '#94a3b8',
                    padding: 12,
                    cornerRadius: 10,
                    displayColors: true,
                    boxPadding: 4,
                    titleFont: { size: 12, weight: 600, family: 'Inter' },
                    bodyFont: { size: 12, family: 'Inter' }
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    grid: {
                        color: 'rgba(255, 255, 255, 0.03)',
                        drawBorder: false
                    },
                    border: { display: false },
                    ticks: {
                        color: '#475569',
                        font: { size: 11, family: 'Inter' },
                        padding: 8,
                        maxTicksLimit: 5
                    }
                },
                x: {
                    grid: { display: false },
                    border: { display: false },
                    ticks: {
                        color: '#475569',
                        font: { size: 11, family: 'Inter' },
                        padding: 8
                    }
                }
            }
        }
    });
}

// ─── Animate Gauges ──────────────────────────────────────────────────────────
function animateGauges() {
    const gauges = document.querySelectorAll('.gauge-progress');
    gauges.forEach(gauge => {
        const targetOffset = gauge.style.strokeDashoffset;
        gauge.style.strokeDashoffset = '440';
        requestAnimationFrame(() => {
            setTimeout(() => {
                gauge.style.strokeDashoffset = targetOffset;
            }, 200);
        });
    });
}

// ─── Counter Animation ───────────────────────────────────────────────────────
function animateCounters() {
    const counters = document.querySelectorAll('.stat-hero-value');
    counters.forEach(el => {
        const target = parseInt(el.textContent, 10);
        if (isNaN(target) || target === 0) return;

        const duration = 1200;
        const start = performance.now();

        function tick(now) {
            const elapsed = now - start;
            const progress = Math.min(elapsed / duration, 1);
            const eased = 1 - Math.pow(1 - progress, 3); // ease-out cubic
            el.textContent = Math.floor(eased * target).toLocaleString();
            if (progress < 1) requestAnimationFrame(tick);
        }

        el.textContent = '0';
        requestAnimationFrame(tick);
    });
}

// ─── Fetch Dashboard Stats ───────────────────────────────────────────────────
async function fetchDashboardStats() {
    try {
        const response = await fetch('/admin/api/stats');
        if (!response.ok) return;

        const data = await response.json();
        updateStat('stat-users', data.users_count || 0);
        updateStat('stat-sessions', data.sessions_count || 0);
        updateStat('stat-roles', data.roles_count || 0);
    } catch (err) {
        // Silently fail — stats are already rendered server-side
    }
}

function updateStat(id, value) {
    const el = document.getElementById(id);
    if (!el) return;

    const current = parseInt(el.textContent, 10) || 0;
    if (current === value) return;

    const duration = 800;
    const start = performance.now();

    function tick(now) {
        const elapsed = now - start;
        const progress = Math.min(elapsed / duration, 1);
        const eased = 1 - Math.pow(1 - progress, 3);
        el.textContent = Math.floor(current + (value - current) * eased).toLocaleString();
        if (progress < 1) requestAnimationFrame(tick);
    }

    requestAnimationFrame(tick);
}
