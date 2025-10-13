const serverAddress = 'http://127.0.0.1:8080/echo'

function sendMessage(event) {
    event.preventDefault();
    const query = document.querySelector('#search').value;
    const results = document.querySelector('#results');

    fetch(serverAddress, {
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
            results.textContent = 'Error: ' + err.message;
            console.error(err);
        });
}

function doSearch(linkString) {
    removeLinks();
    createLinks(linkString);
}

function removeLinks() {
    const links = document.querySelectorAll(".generated-link");

    links.forEach(link => {
        if (link.nextSibling && link.nextSibling.tagName === "BR") {
            link.nextSibling.remove();
        }
        link.remove();
    });
}

function createLinks(linkString) {
    links = linkString.split(" ");
    for (const link of links) {
        if (link !== "") {
            linkEl = document.createElement("a");
            linkEl.setAttribute("href", link);
            linkEl.textContent = link;
            linkEl.classList.add("generated-link");
            document.body.appendChild(linkEl);
            document.body.appendChild(document.createElement("br"));
        }
    }
    document.body.removeChild(document.body.lastElementChild)
}