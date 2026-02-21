/**
 * ABAC Policy Management Scripts (Compiled)
 */

document.addEventListener('DOMContentLoaded', () => {
    if (document.getElementById('policyForm')) {
        initPolicyForm();
    }
});

function initPolicyForm() {
    const form = document.getElementById('policyForm');
    const jsonInput = document.getElementById('conditionsJson');

    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        // Validate JSON
        try {
            JSON.parse(jsonInput.value || '{}');
        } catch (err) {
            showToast('Invalid JSON format in Conditions field', 'warning');
            return;
        }

        const formData = new FormData(form);
        const data = Object.fromEntries(formData.entries());
        data.priority = parseInt(data.priority, 10) || 0;

        const method = data.id ? 'PUT' : 'POST';
        const url = data.id ? `/admin/abac/policies/${data.id}` : '/admin/abac/policies';

        try {
            const response = await fetch(url, {
                method: method,
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(data)
            });

            if (response.ok) {
                window.location.href = '/admin/abac/policies';
            } else {
                const err = await response.json();
                showToast(err.message || 'Failed to save policy', 'error');
            }
        } catch (error) {
            console.error(error);
            showToast('Unexpected error occurred', 'error');
        }
    });
}

window.deletePolicy = async (id) => {
    const confirmed = await showConfirm({
        title: 'Delete Policy',
        message: 'Are you sure you want to delete this policy?',
        confirmText: 'Delete',
        type: 'danger'
    });
    if (!confirmed) return;

    try {
        const response = await fetch(`/admin/abac/policies/${id}`, { method: 'DELETE' });
        if (response.ok) {
            window.location.reload();
        } else {
            showToast('Failed to delete policy', 'error');
        }
    } catch (e) {
        console.error(e);
        showToast('Error deleting policy', 'error');
    }
};
