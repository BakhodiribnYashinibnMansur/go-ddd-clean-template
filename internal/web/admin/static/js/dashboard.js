document.addEventListener('DOMContentLoaded', () => {
    initSparklines();
    initMainAreaChart();
    fetchDashboardStats();
    animateGauges();
});

// Sparkline Charts (Mini line charts)
function initSparklines() {
    const sparklines = document.querySelectorAll('.sparkline');
    sparklines.forEach(canvas => {
        const ctx = canvas.getContext('2d');
        const values = canvas.dataset.values.split(',').map(Number);

        const width = canvas.width = canvas.offsetWidth * 2;
        const height = canvas.height = canvas.offsetHeight * 2;

        const max = Math.max(...values);
        const min = Math.min(...values);
        const range = max - min || 1;

        // Draw gradient fill
        const gradient = ctx.createLinearGradient(0, 0, 0, height);
        gradient.addColorStop(0, 'rgba(99, 102, 241, 0.4)');
        gradient.addColorStop(1, 'rgba(99, 102, 241, 0)');

        ctx.fillStyle = gradient;
        ctx.beginPath();

        const stepX = width / (values.length - 1);
        values.forEach((value, index) => {
            const x = index * stepX;
            const y = height - ((value - min) / range) * height * 0.8 - height * 0.1;
            if (index === 0) {
                ctx.moveTo(x, y);
            } else {
                ctx.lineTo(x, y);
            }
        });

        ctx.lineTo(width, height);
        ctx.lineTo(0, height);
        ctx.closePath();
        ctx.fill();

        // Draw line
        ctx.strokeStyle = '#6366f1';
        ctx.lineWidth = 3;
        ctx.lineJoin = 'round';
        ctx.lineCap = 'round';
        ctx.beginPath();

        values.forEach((value, index) => {
            const x = index * stepX;
            const y = height - ((value - min) / range) * height * 0.8 - height * 0.1;
            if (index === 0) {
                ctx.moveTo(x, y);
            } else {
                ctx.lineTo(x, y);
            }
        });

        ctx.stroke();
    });
}

// Main Area Chart
function initMainAreaChart() {
    const canvas = document.getElementById('mainAreaChart');
    if (!canvas) return;

    new Chart(canvas.getContext('2d'), {
        type: 'line',
        data: {
            labels: ['3298', '2674', '4363', '3956', '4375', '5312', '4865', '9287', '41578'],
            datasets: [{
                label: 'Traffic',
                data: [3200, 2800, 4100, 3800, 4200, 5100, 4700, 5300, 4900],
                borderColor: '#a855f7',
                backgroundColor: function (context) {
                    const ctx = context.chart.ctx;
                    const gradient = ctx.createLinearGradient(0, 0, 0, 320);
                    gradient.addColorStop(0, 'rgba(168, 85, 247, 0.4)');
                    gradient.addColorStop(1, 'rgba(168, 85, 247, 0)');
                    return gradient;
                },
                borderWidth: 3,
                fill: true,
                tension: 0.4,
                pointBackgroundColor: '#a855f7',
                pointBorderColor: '#fff',
                pointBorderWidth: 2,
                pointRadius: 4,
                pointHoverRadius: 6
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: { display: false },
                tooltip: {
                    backgroundColor: 'rgba(15, 23, 42, 0.9)',
                    borderColor: '#a855f7',
                    borderWidth: 1,
                    titleColor: '#f8fafc',
                    bodyColor: '#94a3b8',
                    padding: 12,
                    displayColors: false
                }
            },
            scales: {
                y: {
                    beginAtZero: false,
                    grid: {
                        color: 'rgba(255, 255, 255, 0.05)',
                        drawBorder: false
                    },
                    ticks: {
                        color: '#64748b',
                        font: { size: 11 }
                    }
                },
                x: {
                    grid: { display: false },
                    ticks: {
                        color: '#64748b',
                        font: { size: 11 }
                    }
                }
            }
        }
    });
}

// Animate Gauges on Load
function animateGauges() {
    const gauges = document.querySelectorAll('.gauge-progress');
    gauges.forEach(gauge => {
        const dashOffset = gauge.style.strokeDashoffset;
        gauge.style.strokeDashoffset = '440';
        setTimeout(() => {
            gauge.style.strokeDashoffset = dashOffset;
        }, 100);
    });
}

// Fetch Dashboard Stats from API
async function fetchDashboardStats() {
    try {
        const response = await fetch('/admin/api/stats');
        if (!response.ok) throw new Error('Failed to fetch stats');

        const data = await response.json();

        updateStat('stat-users', data.users_count || 9467);
        updateStat('stat-sessions', data.sessions_count || 3735);
        updateStat('stat-roles', data.roles_count || 2853);

    } catch (error) {
        console.error('Error fetching dashboard stats:', error);
    }
}

function updateStat(id, value) {
    const element = document.getElementById(id);
    if (element) {
        // Animate number counting up
        const start = 0;
        const duration = 1000;
        const startTime = performance.now();

        function animate(currentTime) {
            const elapsed = currentTime - startTime;
            const progress = Math.min(elapsed / duration, 1);
            const current = Math.floor(progress * value);
            element.textContent = current.toLocaleString();

            if (progress < 1) {
                requestAnimationFrame(animate);
            }
        }

        requestAnimationFrame(animate);
    }
}
