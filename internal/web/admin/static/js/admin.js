/**
 * Admin Panel Global Scripts (Compiled JS)
 */
document.addEventListener('DOMContentLoaded', () => {
    initSidebar();
    initAccountModal();
    initGlobalModal();
    initToastContainer();
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
        // Inject backdrop if missing
        if (!modal.querySelector('.modal-backdrop')) {
            const backdrop = document.createElement('div');
            backdrop.className = 'modal-backdrop';
            backdrop.addEventListener('click', () => closeModal(id));
            modal.insertBefore(backdrop, modal.firstChild);
        }
        modal.classList.add('active');
        document.body.style.overflow = 'hidden';
    }
}

function closeModal(id) {
    const modal = document.getElementById(id);
    if (modal) {
        const content = modal.querySelector('.modal-content');
        if (content) {
            content.classList.add('modal-closing');
            setTimeout(() => {
                modal.classList.remove('active');
                content.classList.remove('modal-closing');
                document.body.style.overflow = '';
            }, 180);
        } else {
            modal.classList.remove('active');
            document.body.style.overflow = '';
        }
    }
}

/* ==========================================================================
   Global Confirm / Alert Modal
   ========================================================================== */
let _gmResolve = null;

function initGlobalModal() {
    // Create the modal element once
    const modal = document.createElement('div');
    modal.id = '_globalModal';
    modal.style.cssText = 'display:none; position:fixed; inset:0; z-index:10001; padding:24px; align-items:center; justify-content:center;';
    modal.innerHTML = `
        <div class="modal-backdrop" id="_gmBackdrop"></div>
        <div class="gm-dialog" id="_gmDialog">
            <div class="gm-icon" id="_gmIcon">
                <i class="material-symbols-outlined" id="_gmIconEl">warning</i>
            </div>
            <h3 class="gm-title" id="_gmTitle"></h3>
            <p class="gm-message" id="_gmMessage"></p>
            <div class="gm-actions" id="_gmActions"></div>
        </div>
    `;
    document.body.appendChild(modal);

    // Close on backdrop click
    document.getElementById('_gmBackdrop').addEventListener('click', () => _gmClose(false));

    // Close on Escape
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && document.getElementById('_globalModal').style.display === 'flex') {
            _gmClose(false);
        }
    });
}

function _gmOpen() {
    const modal = document.getElementById('_globalModal');
    const dialog = document.getElementById('_gmDialog');
    modal.style.display = 'flex';
    dialog.classList.remove('modal-closing');
    document.body.style.overflow = 'hidden';
}

function _gmClose(result) {
    const modal = document.getElementById('_globalModal');
    const dialog = document.getElementById('_gmDialog');
    dialog.classList.add('modal-closing');
    setTimeout(() => {
        modal.style.display = 'none';
        dialog.classList.remove('modal-closing');
        document.body.style.overflow = '';
    }, 180);
    if (_gmResolve) {
        _gmResolve(result);
        _gmResolve = null;
    }
}

/**
 * Show a confirm dialog (replaces native confirm())
 * @param {Object} opts - { title, message, confirmText, cancelText, type }
 *   type: 'danger' | 'warning' | 'info' (default: 'danger')
 * @returns {Promise<boolean>}
 */
function showConfirm(opts = {}) {
    // Ensure modal exists (in case DOMContentLoaded hasn't fired yet)
    if (!document.getElementById('_globalModal')) initGlobalModal();

    return new Promise((resolve) => {
        _gmResolve = resolve;
        const type = opts.type || 'danger';
        const icons = { danger: 'warning', warning: 'error', info: 'help' };
        const btnClass = type === 'danger' ? 'gm-btn-danger' : 'gm-btn-primary';

        document.getElementById('_gmIcon').className = 'gm-icon ' + type;
        document.getElementById('_gmIconEl').textContent = icons[type] || 'warning';
        document.getElementById('_gmTitle').textContent = opts.title || 'Confirm';
        document.getElementById('_gmMessage').textContent = opts.message || 'Are you sure?';
        document.getElementById('_gmActions').innerHTML = `
            <button class="gm-btn gm-btn-cancel" onclick="_gmClose(false)">${opts.cancelText || 'Cancel'}</button>
            <button class="gm-btn ${btnClass}" onclick="_gmClose(true)">${opts.confirmText || 'Confirm'}</button>
        `;
        _gmOpen();
    });
}

/**
 * Show an alert dialog (replaces native alert())
 * @param {Object} opts - { title, message, buttonText, type }
 *   type: 'danger' | 'warning' | 'info' | 'success' (default: 'info')
 * @returns {Promise<void>}
 */
function showAlert(opts = {}) {
    if (typeof opts === 'string') opts = { message: opts };
    // Ensure modal exists
    if (!document.getElementById('_globalModal')) initGlobalModal();
    return new Promise((resolve) => {
        _gmResolve = () => resolve();
        const type = opts.type || 'info';
        const icons = { danger: 'error', warning: 'warning', info: 'info', success: 'check_circle' };

        document.getElementById('_gmIcon').className = 'gm-icon ' + type;
        document.getElementById('_gmIconEl').textContent = icons[type] || 'info';
        document.getElementById('_gmTitle').textContent = opts.title || (type === 'danger' ? 'Error' : type === 'success' ? 'Success' : 'Notice');
        document.getElementById('_gmMessage').textContent = opts.message || '';
        document.getElementById('_gmActions').innerHTML = `
            <button class="gm-btn gm-btn-primary" onclick="_gmClose(true)">${opts.buttonText || 'OK'}</button>
        `;
        _gmOpen();
    });
}

/* ==========================================================================
   Toast Notification System
   ========================================================================== */
function initToastContainer() {
    if (!document.getElementById('_toastContainer')) {
        const container = document.createElement('div');
        container.id = '_toastContainer';
        container.className = 'toast-container';
        document.body.appendChild(container);
    }
}

/**
 * Show a toast notification (replaces native alert() for non-blocking messages)
 * @param {string} message
 * @param {string} type - 'error' | 'success' | 'warning' | 'info' (default: 'info')
 * @param {number} duration - ms (default: 4000)
 */
function showToast(message, type = 'info', duration = 4000) {
    const container = document.getElementById('_toastContainer');
    if (!container) return;

    const icons = {
        error: 'error',
        success: 'check_circle',
        warning: 'warning',
        info: 'info'
    };

    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.innerHTML = `
        <span class="toast-icon"><i class="material-symbols-outlined">${icons[type] || 'info'}</i></span>
        <span class="toast-text">${message}</span>
        <button class="toast-close" onclick="this.parentElement.remove()"><i class="material-symbols-outlined">close</i></button>
    `;
    container.appendChild(toast);

    setTimeout(() => {
        toast.classList.add('toast-out');
        setTimeout(() => toast.remove(), 200);
    }, duration);
}

/* ==========================================================================
   Override native alert() to use custom modal
   ========================================================================== */
(function() {
    const _nativeAlert = window.alert;
    window.alert = function(message) {
        try {
            showAlert({ message: String(message), type: 'info' });
        } catch (_) {
            _nativeAlert.call(window, message);
        }
    };
})();
