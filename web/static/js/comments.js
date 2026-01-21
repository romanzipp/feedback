document.addEventListener('DOMContentLoaded', function() {
    const forms = document.querySelectorAll('.comment-form');

    forms.forEach(form => {
        form.addEventListener('submit', async function(e) {
            e.preventDefault();

            const fileId = this.dataset.fileId;
            const content = this.querySelector('[name="content"]').value;
            const commentsContainer = document.getElementById(`comments-${fileId}`);

            try {
                const response = await fetch(`/api/files/${fileId}/comments`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/x-www-form-urlencoded',
                    },
                    body: `content=${encodeURIComponent(content)}`
                });

                if (!response.ok) {
                    throw new Error('Failed to post comment');
                }

                const comment = await response.json();

                // Add comment to UI
                const commentDiv = document.createElement('div');
                commentDiv.className = 'bg-gray-50 rounded p-3';
                commentDiv.innerHTML = `
                    <div class="flex items-baseline gap-2 mb-1">
                        <span class="font-medium text-gray-900">${escapeHtml(comment.Username)}</span>
                        <span class="text-sm text-gray-500">${formatDate(comment.CreatedAt)}</span>
                    </div>
                    <p class="text-gray-700">${escapeHtml(comment.Content)}</p>
                `;
                commentsContainer.appendChild(commentDiv);

                // Clear form
                this.reset();

            } catch (error) {
                alert('Failed to post comment. Please try again.');
                console.error(error);
            }
        });
    });
});

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString('en-US', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
    });
}
