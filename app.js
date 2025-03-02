
document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('scrapeForm');
    const loadingElement = document.createElement('div');
    loadingElement.id = 'loading';
    loadingElement.innerHTML = 'Loading... <div class="spinner"></div>';
    loadingElement.style.display = 'none';
    document.querySelector('.container').appendChild(loadingElement);

    form.addEventListener('submit', function(e) {
        e.preventDefault();

        const url = document.getElementById('url').value.trim();
        if (!url) {
            showError('Please enter a URL');
            return;
        }

        // Show loading indicator
        loadingElement.style.display = 'block';
        
        // Create and open a new tab for downloading Excel file
        const downloadUrl = `/scrape?url=${encodeURIComponent(url)}`;
        window.open(downloadUrl, '_blank');

        // Make a request to get JSON data for display on the page
        fetch(`/scrape?url=${encodeURIComponent(url)}`, {
            headers: {
                'Accept': 'application/json'
            }
        })
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! Status: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            // Hide loading indicator
            loadingElement.style.display = 'none';
            
            // Display results
            displayResults(data);
        })
        .catch(error => {
            // Hide loading indicator
            loadingElement.style.display = 'none';
            
            console.error('Error:', error);
            showError('An error occurred while fetching data: ' + error.message);
        });
    });

    function displayResults(data) {
        let resultsContainer = document.getElementById('resultsContainer');
        if (!resultsContainer) {
            // Create container if it doesn't exist yet
            const container = document.querySelector('.container');
            resultsContainer = document.createElement('div');
            resultsContainer.id = 'resultsContainer';
            container.appendChild(resultsContainer);
            
            const resultsTitle = document.createElement('h2');
            resultsTitle.textContent = 'Extracted Data';
            resultsContainer.appendChild(resultsTitle);
        } else {
            resultsContainer.innerHTML = '';
            const resultsTitle = document.createElement('h2');
            resultsTitle.textContent = 'Extracted Data';
            resultsContainer.appendChild(resultsTitle);
        }
        
        if (data.length === 0) {
            const noData = document.createElement('p');
            noData.className = 'no-data';
            noData.textContent = 'No data found';
            resultsContainer.appendChild(noData);
            return;
        }

        // Group data by type
        const groupedData = {};
        data.forEach(item => {
            if (!groupedData[item.type]) {
                groupedData[item.type] = [];
            }
            groupedData[item.type].push(item);
        });

        // Create a table to display structured data
        const table = document.createElement('table');
        table.id = 'data-table';
        
        // Create table header
        const thead = document.createElement('thead');
        const headerRow = document.createElement('tr');
        const headers = ['Type', 'Tag', 'Content', 'URL'];
        
        headers.forEach(headerText => {
            const th = document.createElement('th');
            th.textContent = headerText;
            headerRow.appendChild(th);
        });
        
        thead.appendChild(headerRow);
        table.appendChild(thead);
        
        // Create table body
        const tbody = document.createElement('tbody');
        
        // Add rows for each type of content
        Object.keys(groupedData).forEach(type => {
            // Add section header
            const sectionRow = document.createElement('tr');
            sectionRow.className = 'section-header';
            
            const sectionCell = document.createElement('td');
            sectionCell.colSpan = headers.length;
            sectionCell.textContent = type.charAt(0).toUpperCase() + type.slice(1) + 's';
            sectionRow.appendChild(sectionCell);
            tbody.appendChild(sectionRow);
            
            // Add data rows
            groupedData[type].forEach(item => {
                const dataRow = document.createElement('tr');
                
                const typeCell = document.createElement('td');
                typeCell.textContent = item.type;
                dataRow.appendChild(typeCell);
                
                const tagCell = document.createElement('td');
                tagCell.textContent = item.tag;
                dataRow.appendChild(tagCell);
                
                const contentCell = document.createElement('td');
                contentCell.textContent = item.text;
                dataRow.appendChild(contentCell);
                
                const urlCell = document.createElement('td');
                if (item.url) {
                    const link = document.createElement('a');
                    link.href = item.url;
                    link.textContent = item.url;
                    link.target = '_blank';
                    urlCell.appendChild(link);
                }
                dataRow.appendChild(urlCell);
                
                tbody.appendChild(dataRow);
            });
        });
        
        table.appendChild(tbody);
        resultsContainer.appendChild(table);
        
        // Show summary
        const summary = document.createElement('div');
        summary.className = 'summary';
        
        const counts = Object.keys(groupedData).map(type => {
            return `${groupedData[type].length} ${type}s`;
        }).join(', ');
        
        summary.textContent = `Found: ${counts}`;
        resultsContainer.appendChild(summary);
    }

    function showError(message) {
        const errorContainer = document.getElementById('error') || document.createElement('div');
        if (!errorContainer.id) {
            errorContainer.id = 'error';
            document.querySelector('.container').insertBefore(errorContainer, form.nextSibling);
        }
        errorContainer.textContent = message;
        errorContainer.style.display = 'block';
        
        // Automatically hide message after 5 seconds
        setTimeout(() => {
            errorContainer.style.display = 'none';
        }, 5000);
    }
});
