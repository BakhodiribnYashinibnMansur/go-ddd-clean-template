/**
 * Dashboard Scripts
 */
document.addEventListener('DOMContentLoaded', () => {
    initDashboardCharts();
    fetchDashboardStats();
});

function initDashboardCharts() {
    // 1. User Growth (Line Chart)
    const userCanvas = document.getElementById('userGrowthChart');
    if (userCanvas) {
        new Chart(userCanvas.getContext('2d'), {
            type: 'line',
            data: {
                labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
                datasets: [{
                    label: 'New Users',
                    data: [12, 19, 3, 5, 2, 3],
                    borderColor: '#4f46e5',
                    backgroundColor: 'rgba(79, 70, 229, 0.1)',
                    tension: 0.4,
                    fill: true
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { display: false }
                },
                scales: {
                    x: { grid: { display: false, color: 'rgba(255, 255, 255, 0.05)' } },
                    y: { grid: { color: 'rgba(255, 255, 255, 0.05)' } }
                }
            }
        });
    }

    // 2. Traffic Sources (Doughnut)
    const trafficCanvas = document.getElementById('trafficSourceChart');
    if (trafficCanvas) {
        new Chart(trafficCanvas.getContext('2d'), {
            type: 'doughnut',
            data: {
                labels: ['Direct', 'Social', 'Referral'],
                datasets: [{
                    data: [55, 30, 15],
                    backgroundColor: ['#4f46e5', '#10b981', '#f59e0b'],
                    borderWidth: 0
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { position: 'bottom' }
                }
            }
        });
    }
}

async function fetchDashboardStats() {
    try {
        const response = await fetch('/admin/api/stats');
        if (!response.ok) {
            throw new Error('Failed to fetch stats');
        }
        const data = await response.json();

        updateStat('stat-users', data.users_count);
        updateStat('stat-sessions', data.sessions_count);
        updateStat('stat-roles', data.roles_count);
        updateStat('stat-perms', data.perms_count);
        updateStat('stat-scopes', data.scopes_count);
        updateStat('stat-policies', data.policies_count);

    } catch (error) {
        console.error('Error fetching dashboard stats:', error);
    }
}

function updateStat(id, value) {
    const element = document.getElementById(id);
    if (element) {
        element.textContent = value;
    }
}
