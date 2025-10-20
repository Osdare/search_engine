const serverAddress = 'https://api.orb.ax';
//const serverAddress = 'http://127.0.0.1:8080';
const linksEndpoint = '/links';
const imagesEndpoint = '/images';

document.addEventListener('DOMContentLoaded', () => {
  document.querySelector('#searchForm').addEventListener('submit', handleSearch);
});

function handleSearch(event) {
  event.preventDefault();
  const query = document.querySelector('#search').value.trim();
  const searchType = document.querySelector('#searchType').value;

  if (!query) return;

  if (searchType === 'links') {
    sendMessage(query);
  } else {
    sendImageMessage(query);
  }
}

function sendMessage(query) {
  fetch(serverAddress + linksEndpoint, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: 'message=' + encodeURIComponent(query),
  })
    .then((res) => {
      if (!res.ok) throw new Error('Server error');
      return res.text();
    })
    .then((text) => {
      doSearch(text);
    })
    .catch((err) => {
      showError(err.message);
    });
}

function sendImageMessage(query) {
  fetch(serverAddress + imagesEndpoint, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: 'message=' + encodeURIComponent(query),
  })
    .then((res) => {
      if (!res.ok) throw new Error('Server error');
      return res.text();
    })
    .then((text) => {
      createImages(text);
    })
    .catch((err) => {
      showError(err.message);
    });
}

function doSearch(linkString) {
  clearResults();
  createLinks(linkString);
}

function createLinks(linkString) {
  const links = linkString.split(' ').filter(Boolean);
  const resultsDiv = document.querySelector('#results');

  for (const link of links) {
    const linkEl = document.createElement('a');
    linkEl.href = link;
    linkEl.textContent = link;
    linkEl.classList.add('generated-link');
    linkEl.target = '_blank';
    resultsDiv.appendChild(linkEl);
    resultsDiv.appendChild(document.createElement('br'));
    resultsDiv.appendChild(document.createElement('br'));
  }
}

function createImages(imageString) {
  clearResults();
  const imageLinks = imageString.split(' ').filter(Boolean);
  const resultsDiv = document.querySelector('#results');

  for (const link of imageLinks) {
    const imgEl = document.createElement('img');
    imgEl.src = link;
    imgEl.alt = link;
    imgEl.classList.add('generated-image');
    imgEl.style.maxWidth = '200px';
    imgEl.style.margin = '5px';
    resultsDiv.appendChild(imgEl);
  }
}

function clearResults() {
  document.querySelector('#results').innerHTML = '';
}

function showError(msg) {
  const resultsDiv = document.querySelector('#results');
  resultsDiv.textContent = 'Error: ' + msg;
}
