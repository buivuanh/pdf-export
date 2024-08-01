document.addEventListener('DOMContentLoaded', function () {
    /* Ensure maximum 5 block per page */
    var blocks = document.querySelectorAll('.block');
    for (var i = 0; i < blocks.length; i++) {
        if ((i + 1) % 5 === 0) {
            blocks[i].style.pageBreakAfter = 'always';
        }
    }
});