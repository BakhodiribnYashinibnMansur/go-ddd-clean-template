/**
 * RBAC Management Scripts
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
                alert('Error: ' + (err.message || 'Failed to save role'));
            }
        } catch (error) {
            console.error(error);
            alert('Unexpected error occurred');
        }
    });
}

// Exposed to global scope for button onclick attributes
window.deleteRole = async (id) => {
    if (!confirm('Are you sure you want to delete this role? This might affect assigned users.')) return;

    try {
        const response = await fetch(`/admin/rbac/roles/${id}`, { method: 'DELETE' });
        if (response.ok) {
            window.location.reload();
        } else {
            alert('Failed to delete role');
        }
    } catch (e) {
        console.error(e);
        alert('Error deleting role');
    }
};
