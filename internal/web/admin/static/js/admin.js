/**
 * Admin Panel Global Scripts (Compiled JS)
 */
document.addEventListener('DOMContentLoaded', () => {
    initSidebar();
    initAccountModal();
});

// Smart header removed


// Smart header removed

function initSidebar() {
    const layout = document.querySelector('.admin-layout');
    const sidebar = document.querySelector('.sidebar');
    const collapseBtn = document.getElementById('sidebarCollapseBtn');

    // Restore state
    const savedState = localStorage.getItem('sidebar-collapsed');
    if (savedState === 'true' && layout) {
        layout.classList.add('collapsed');
    }

    const toggleSidebar = () => {
        if (layout) {
            layout.classList.toggle('collapsed');
            const isCollapsed = layout.classList.contains('collapsed');
            localStorage.setItem('sidebar-collapsed', String(isCollapsed));

            // Remove hover class if manually checking
            if (!isCollapsed && layout.classList.contains('hover-expanded')) {
                layout.classList.remove('hover-expanded');
            }
        }
    };

    if (collapseBtn) {
        collapseBtn.addEventListener('click', toggleSidebar);
    }

    // Auto-expand on hover (Amazing Feature)
    if (sidebar && layout) {
        sidebar.addEventListener('mouseenter', () => {
            if (layout.classList.contains('collapsed')) {
                layout.classList.add('hover-expanded');
            }
        });

        sidebar.addEventListener('mouseleave', () => {
            if (layout.classList.contains('hover-expanded')) {
                layout.classList.remove('hover-expanded');
            }
        });
    }
    // Auto-inject tooltips for collapsed mode
    const navItems = document.querySelectorAll('.nav-item');
    navItems.forEach(item => {
        const text = item.querySelector('.nav-text')?.textContent;
        if (text) {
            item.setAttribute('title', text.trim());
        }
    });
}

function initAccountModal() {
    const modal = document.getElementById('accountModal');
    const openBtns = document.querySelectorAll('.profile-trigger');
    const closeBtn = document.getElementById('closeAccountModal');

    if (modal && openBtns.length > 0) {
        openBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.preventDefault();
                modal.classList.add('active');
            });
        });
    }

    if (modal && closeBtn) {
        closeBtn.addEventListener('click', () => {
            modal.classList.remove('active');
        });
    }

    if (modal) {
        // Close on outside click
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.classList.remove('active');
            }
        });
    }
}

// Global Utilities
function filterTable(input) {
    const filter = input.value.toUpperCase();
    const table = document.querySelector('.data-table') || document.querySelector('.table');
    if (!table) return;

    const tr = table.getElementsByTagName('tr');
    for (let i = 1; i < tr.length; i++) { // Skip header
        let rowContent = tr[i].textContent || tr[i].innerText;
        if (rowContent.toUpperCase().indexOf(filter) > -1) {
            tr[i].style.display = "";
        } else {
            tr[i].style.display = "none";
        }
    }
}

function openModal(id) {
    const modal = document.getElementById(id);
    if (modal) {
        modal.classList.add('active');
        document.body.style.overflow = 'hidden'; // Prevent background scroll
    }
}

function closeModal(id) {
    const modal = document.getElementById(id);
    if (modal) {
        modal.classList.remove('active');
        document.body.style.overflow = '';
    }
}
