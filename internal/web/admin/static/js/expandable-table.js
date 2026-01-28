// Expandable Table/List Component - Fixed Version
document.addEventListener('DOMContentLoaded', function () {
    console.log('Expandable table initialized');

    // Find all table rows
    const tableRows = document.querySelectorAll('.table-row');
    console.log('Found rows:', tableRows.length);

    tableRows.forEach(function (row) {
        row.addEventListener('click', function (e) {
            console.log('Row clicked');

            // Don't toggle if clicking on a link or button inside (except expand-btn)
            if (e.target.closest('a') || (e.target.closest('button') && !e.target.closest('.expand-btn'))) {
                return;
            }

            // Get the wrapper and details
            const wrapper = row.closest('.table-row-wrapper');
            const details = wrapper.querySelector('.row-details');

            // Toggle classes
            row.classList.toggle('expanded');
            details.classList.toggle('open');

            console.log('Toggled - expanded:', row.classList.contains('expanded'));
        });
    });

    // Also add click handler to expand buttons specifically
    const expandBtns = document.querySelectorAll('.expand-btn');
    console.log('Found expand buttons:', expandBtns.length);

    expandBtns.forEach(function (btn) {
        btn.addEventListener('click', function (e) {
            e.stopPropagation(); // Prevent double trigger
            console.log('Expand button clicked');

            const row = btn.closest('.table-row');
            const wrapper = row.closest('.table-row-wrapper');
            const details = wrapper.querySelector('.row-details');

            // Toggle classes
            row.classList.toggle('expanded');
            details.classList.toggle('open');
        });
    });
});
