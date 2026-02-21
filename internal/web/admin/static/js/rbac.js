/**
 * RBAC Management Scripts (Compiled)
 */

document.addEventListener('DOMContentLoaded', () => {
    // Only init if we are on an RBAC page
    if (document.getElementById('roleForm')) {
        initRoleForm();
    }
});

function initRoleForm() {
    const form = document.getElementById('roleForm');
    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        const formData = new FormData(form);
        const data = Object.fromEntries(formData.entries());

        // Handle multi-select checkboxes for permissions
        const checkboxes = form.querySelectorAll('input[name="permissions"]:checked');
        data.permissions = Array.from(checkboxes).map((cb) => cb.value);

        const method = data.id ? 'PUT' : 'POST';
        const url = data.id ? `/admin/rbac/roles/${data.id}` : '/admin/rbac/roles';

        try {
            const response = await fetch(url, {
                method: method,
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(data)
            });

            if (response.ok) {
                window.location.href = '/admin/rbac/roles';
            } else {
                const err = await response.json();
                showToast(err.message || 'Failed to save role', 'error');
            }
        } catch (error) {
            console.error(error);
            showToast('Unexpected error occurred', 'error');
        }
    });
}

// Exposed to global scope for button onclick attributes
window.deleteRole = async (id) => {
    const confirmed = await showConfirm({
        title: 'Delete Role',
        message: 'Are you sure you want to delete this role? This might affect assigned users.',
        confirmText: 'Delete',
        type: 'danger'
    });
    if (!confirmed) return;

    try {
        const response = await fetch(`/admin/rbac/roles/${id}`, { method: 'DELETE' });
        if (response.ok) {
            window.location.reload();
        } else {
            showToast('Failed to delete role', 'error');
        }
    } catch (e) {
        console.error(e);
        showToast('Error deleting role', 'error');
    }
};
