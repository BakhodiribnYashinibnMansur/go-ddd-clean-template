/**
 * Admin Panel Global Scripts
 */
document.addEventListener('DOMContentLoaded', () => {
    initSidebar();
    initLogout();
});

function initSidebar(): void {
    // Basic sidebar toggle logic for mobile
    const toggleBtn = document.getElementById('sidebarToggle');
    const sidebar = document.getElementById('sidebar');

    if (toggleBtn && sidebar) {
        toggleBtn.addEventListener('click', () => {
            sidebar.classList.toggle('open');
        });
    }
}

function initLogout(): void {
    const logoutBtn = document.getElementById('logoutBtn');
    if (!logoutBtn) return;

    logoutBtn.addEventListener('click', (e: MouseEvent) => {
        e.preventDefault();
        if (confirm('Are you sure you want to logout?')) {
            // Ideally this should be a POST request to prevent CSRF on GET
            // Creating a form dynamically
            const form = document.createElement('form');
            form.method = 'POST';
            form.action = '/admin/logout';

            // Add CSRF token if available globally
            const csrfMeta = document.querySelector('meta[name="csrf-token"]') as HTMLMetaElement | null;
            if (csrfMeta && csrfMeta.content) {
                const input = document.createElement('input');
                input.type = 'hidden';
                input.name = '_csrf';
                input.value = csrfMeta.content;
                form.appendChild(input);
            }

            document.body.appendChild(form);
            form.submit();
        }
    });
}
