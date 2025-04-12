document.addEventListener('DOMContentLoaded', function() {
    const emailList = document.getElementById('emailList');
    const emailContent = document.getElementById('emailContent');
    const clearBtn = document.getElementById('clearBtn');

    let activeEmailId = null;
    
    // Load the list of emails
    function loadEmails() {
        fetch('/api/emails')
            .then(response => response.json())
            .then(emails => {
                if (emails.length === 0) {
                    emailList.innerHTML = '<div class="placeholder"><p>No emails received</p></div>';
                    emailContent.innerHTML = '<div class="placeholder"><p>No emails to display</p></div>';
                    return;
                }
                
                emailList.innerHTML = '';
                emails.forEach(email => {
                    const item = document.createElement('div');
                    item.className = 'email-item';
                    if (email.ID === activeEmailId) {
                        item.classList.add('active');
                    }
                    
                    const date = new Date(email.timestamp);
                    const formattedDate = date.toLocaleString();
                    
                    item.innerHTML = '<h3>' + (email.subject || '(No subject)') + '</h3>' +
                        '<div class="email-meta">' +
                        '<span>From: ' + email.from + '</span>' +
                        '<span>' + formattedDate + '</span>' +
                        '</div>';
                    
                    item.addEventListener('click', () => {
                        document.querySelectorAll('.email-item').forEach(el => el.classList.remove('active'));
                        item.classList.add('active');
                        loadEmailContent(email.ID);
                    });
                    
                    emailList.appendChild(item);
                });
                
                // Load the first email if none is selected
                if (!activeEmailId && emails.length > 0) {
                    loadEmailContent(emails[0].ID);
                }
            })
            .catch(error => {
                console.error('Error loading emails:', error);
                emailList.innerHTML = '<div class="placeholder"><p>Error loading emails</p></div>';
            });
    }
    
    // Load the content of a specific email
    function loadEmailContent(id) {
        activeEmailId = id;
        
        fetch('/api/emails/' + id)
            .then(response => response.json())
            .then(email => {
                const date = new Date(email.timestamp);
                const formattedDate = date.toLocaleString();
                
                emailContent.innerHTML = 
                    '<div class="email-header">' +
                    '<h2>' + (email.subject || '(No subject)') + '</h2>' +
                    '<div class="email-details"><strong>From:</strong> ' + email.from + '</div>' +
                    '<div class="email-details"><strong>To:</strong> ' + email.to.join(', ') + '</div>' +
                    '<div class="email-details"><strong>Date:</strong> ' + formattedDate + '</div>' +
                    '</div>' +
                    '<div class="email-body">' +
                    (email.html ? email.body : '<pre>' + email.body + '</pre>') +
                    '</div>';
            })
            .catch(error => {
                console.error('Error loading email content:', error);
                emailContent.innerHTML = '<div class="placeholder"><p>Error loading email content</p></div>';
            });
    }
    
    // Clear all emails
    clearBtn.addEventListener('click', function() {
        if (confirm('Are you sure you want to delete all emails?')) {
            fetch('/api/clear', { method: 'POST' })
                .then(response => {
                    if (response.ok) {
                        activeEmailId = null;
                        loadEmails();
                    }
                })
                .catch(error => console.error('Error clearing emails:', error));
        }
    });
    
    // Load emails initially
    loadEmails();
    
    // Update the list of emails every 10 seconds
    setInterval(loadEmails, 10000);
}); 