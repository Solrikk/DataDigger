document.getElementById('scrapeForm').addEventListener('submit', function(e) {
    e.preventDefault();

    const url = document.getElementById('url').value;
    window.open(`/scrape?url=${encodeURIComponent(url)}`, '_blank');

    fetch(`/scrape?url=${encodeURIComponent(url)}`)
        .then(response => response.json())
        .then(data => {
            const results = document.getElementById('results');
            results.innerHTML = '';
            data.forEach(item => {
                const li = document.createElement('li');
                if (item.text) {
                    li.textContent = item.text;
                } else if (item.url) {
                    const a = document.createElement('a');
                    a.href = item.url;
                    a.textContent = item.url;
                    li.appendChild(a);
                }
                results.appendChild(li);
            });
        })
        .catch(error => {
            console.error('Error:', error);
        });
});