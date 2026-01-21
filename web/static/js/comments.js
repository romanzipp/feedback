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
                commentDiv.className = 'bg-gray-50 rounded p-2';
                const commentDate = new Date(comment.CreatedAt);
                commentDiv.innerHTML = `
                    <div class="flex items-baseline gap-1 mb-1">
                        <span class="text-xs font-medium text-gray-900">${escapeHtml(comment.Username)}</span>
                        <span class="text-xs text-gray-400 relative-time" data-time="${commentDate.toISOString()}">${getRelativeTime(commentDate)}</span>
                    </div>
                    <p class="text-xs text-gray-700">${escapeHtml(comment.Content)}</p>
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

function getRelativeTime(date) {
    const seconds = Math.floor((new Date() - date) / 1000);

    if (seconds < 60) return 'just now';
    if (seconds < 3600) {
        const mins = Math.floor(seconds / 60);
        return mins + 'm ago';
    }
    if (seconds < 86400) {
        const hours = Math.floor(seconds / 3600);
        return hours + 'h ago';
    }
    if (seconds < 2592000) {
        const days = Math.floor(seconds / 86400);
        return days + 'd ago';
    }

    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return month + '/' + day;
}
