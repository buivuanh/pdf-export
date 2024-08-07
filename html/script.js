document.addEventListener('DOMContentLoaded', function () {
    /* Ensure maximum 5 block per page */
    var blocks = document.querySelectorAll('.block');
    for (var i = 0; i < blocks.length; i++) {
        if ((i + 1) % 5 === 0) {
            blocks[i].style.pageBreakAfter = 'always';
        }
    }
    /* Ensure maximum 5 row table per page */
    var blocks2 = document.querySelectorAll('.table-section tbody tr');
    for (var i = 0; i < blocks2.length; i++) {
        if ((i + 1) % 5 === 0) {
            blocks2[i].style.pageBreakAfter = 'always';
            // blocks2[i].style.paddingTop = '50px';
        }
    }
});

// Function to adjust font size of agent name based on its length
function adjustAgentNameFontSize() {
    var agentNameElement = document.querySelector('.agent-name');
    var agentNameLength = agentNameElement.textContent.length;
    if (agentNameLength > 16) { // Adjust this threshold as needed
        agentNameElement.style.fontSize = '42px';
        agentNameElement.style.lineHeight = '47px';
        agentNameElement.style.maxHeight = '94px';
    }
}

// Call the function when the page loads
window.onload = function() {
    adjustAgentNameFontSize();
};

// Call the function whenever the window is resized (optional)
window.onresize = function() {
    adjustAgentNameFontSize();
};