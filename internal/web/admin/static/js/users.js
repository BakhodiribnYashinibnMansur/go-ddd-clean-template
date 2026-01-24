/**
 * Users Management Scripts (Compiled)
 */
document.addEventListener('DOMContentLoaded', () => {
    initUserActions();
    initBulkSelection();
});

function initUserActions() {
    const blockBtns = document.querySelectorAll('.block-user-btn');
    const unblockBtns = document.querySelectorAll('.unblock-user-btn');

    blockBtns.forEach((btn) => {
        btn.addEventListener('click', () => {
            const id = btn.dataset.id;
            if (id) handleUserAction(id, 'block');
        });
    });

    unblockBtns.forEach((btn) => {
        btn.addEventListener('click', () => {
            const id = btn.dataset.id;
            if (id) handleUserAction(id, 'unblock');
        });
    });
}

function initBulkSelection() {
    const selectAll = document.getElementById('selectAll');
    if (!selectAll) return;

    const checkboxes = document.querySelectorAll('.user-checkbox');
    const bulkActions = document.getElementById('bulkActions');
    const selectedCount = document.getElementById('selectedCount');

    function updateState() {
        const checked = document.querySelectorAll('.user-checkbox:checked');
        const count = checked.length;

        if (bulkActions) {
            if (count > 0) {
                bulkActions.style.display = 'flex';
                if (selectedCount) selectedCount.innerText = `${count} selected`;
            } else {
                bulkActions.style.display = 'none';
            }
        }
    }

    selectAll.addEventListener('change', () => {
        checkboxes.forEach(cb => cb.checked = selectAll.checked);
        updateState();
    });

    checkboxes.forEach(cb => {
        cb.addEventListener('change', () => {
            if (!cb.checked) selectAll.checked = false;
            updateState();
        });
    });
}

async function handleUserAction(userId, action) {
    if (!userId) return;

    const actionText = action === 'block' ? 'Block' : 'Unblock';
    const confirmMessage = `Are you sure you want to ${actionText} user #${userId}?`;

    if (!confirm(confirmMessage)) {
        return;
    }

    try {
        const response = await fetch(`/admin/users/${userId}/${action}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            }
        });

        if (response.ok) {
            window.location.reload();
        } else {
            const data = await response.json();
            alert(`Error: ${data.message || 'Action failed'}`);
        }
    } catch (error) {
        console.error('Error performing user action:', error);
        alert('An unexpected error occurred.');
    }
}

async function bulkAction(action) {
    const checked = document.querySelectorAll('.user-checkbox:checked');
    if (checked.length === 0) return;

    const ids = Array.from(checked).map(cb => cb.value);

    if (!confirm(`Are you sure you want to ${action} ${ids.length} users?`)) {
        return;
    }

    try {
        const response = await fetch('/admin/users/bulk-action', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ ids: ids, action: action })
        });

        if (response.ok) {
            window.location.reload();
        } else {
            alert('Bulk action failed');
        }
    } catch (e) {
        console.error(e);
        alert('Error performing bulk action');
    }
}

// Make globally available if needed for onclick handlers
window.bulkAction = bulkAction;

async function deleteUser(id) {
    if (!confirm('Are you sure you want to delete this user? This action cannot be undone.')) return;

    try {
        const response = await fetch('/admin/users/bulk-action', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ ids: [id], action: 'delete' })
        });

        if (response.ok) {
            window.location.reload();
        } else {
            alert('Delete failed');
        }
    } catch (e) {
        console.error(e);
        alert('Error performing delete');
    }
}
window.deleteUser = deleteUser;
