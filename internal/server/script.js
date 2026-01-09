// hold downloads from previous fetch
let prevDownloads = [];

// send request to stop / resume download
async function toggleDownload(id) {
  try {
    const resp = await fetch(`/api/toggle/${id}`, {
      method: "GET",
      credentials: "include",
    });

    if (resp.ok) {
      getDownloadsAndFillTable();
    } else {
      console.log(resp);
    }
  } catch (err) {
    console.error("Fetch failed:", err);
  }
}

//send request to delete download
async function deleteDownload(id) {
  try {
    const resp = await fetch(`/api/delete/${id}`, {
      method: "GET",
      credentials: "include",
    });

    if (resp.ok) {
      return true;
    } else {
      console.log(resp);
    }
  } catch (err) {
    console.error("Fetch failed:", err);
  }
}

// format bytes to human readable and keep 2 decimal points
function formatBytes(bytes) {
  if (bytes > 1_000_000_000) {
    return `${(bytes / 1_000_000_000).toFixed(2)} GB`;
  } else if (bytes > 1_000_000) {
    return `${(bytes / 1_000_000).toFixed(2)} MB`;
  } else if (bytes > 1_000) {
    return `${(bytes / 1_000).toFixed(2)} KB`;
  } else {
    return `${bytes.toFixed(2)} B`;
  }
}

// fill table with data
function fillTable(downloads) {
  // disable table if no items in DB
  if (downloads.length === 0) {
    const table = (document.getElementById("downloadsTable").style.display =
      "none");
  } else {
    const table = (document.getElementById("downloadsTable").style.display =
      "block");
  }
  console.log(downloads);
  const tbody = document.getElementById("table-body");

  downloads.forEach((d) => {
    const rowId = `row${d.id}`;
    let row = document.getElementById(rowId);

    // row is not present create
    if (!row) {
      row = document.createElement("tr");
      row.id = rowId;

      for (let i = 0; i < 8; i++) {
        const td = document.createElement("td");
        row.appendChild(td);
      }

      row.cells[5].textContent = "0 MB/s";

      toggleBtn = document.createElement("button");
      toggleBtn.textContent = "Toggle";
      toggleBtn.classList.add("buttonBlue");
      toggleBtn.addEventListener("click", () => {
        toggleDownload(d.id);
      });

      deleteBtn = document.createElement("button");
      deleteBtn.textContent = "Delete";
      deleteBtn.classList.add("buttonRed");
      deleteBtn.addEventListener("click", () => {
        if (deleteDownload(d.id)) {
          row.remove();
        }
      });

      row.cells[6].appendChild(toggleBtn);
      row.cells[7].appendChild(deleteBtn);

      row.cells[0].textContent = d.id;
      row.cells[2].textContent = d.filename;
      row.cells[3].textContent = `0.0 GB`;

      tbody.appendChild(row);

      // row is present just change calculated data
    } else {
      const prevDownload = prevDownloads.find(
        (download) => download.id === d.id,
      );

      row.cells[5].textContent =
        formatBytes((d.downloaded - prevDownload.downloaded) / 2) + "/s";
    }

    let status;
    if (d.active) {
      status = "active";
    } else if (d.completed) {
      status = "finished";
    } else if (!d.active && !d.completed && d.err === "") {
      status = "stopped";
    } else {
      status = d.err;
    }

    row.cells[1].textContent = status;
    row.cells[3].textContent = formatBytes(d.downloaded);
    row.cells[4].textContent = formatBytes(d.size);
  });
  prevDownloads = downloads;
}

// get free space and name of current dir
async function getDirInfo() {
  try {
    const resp = await fetch("/api/info", {
      method: "GET",
      credentials: "include",
    });

    if (resp.ok) {
      const dirInfo = await resp.json();
      console.log(dirInfo);
      document.getElementById("freeSpace").innerHTML += " " + dirInfo.freeSpace;
      document.getElementById("dirPathData").innerText = " " + dirInfo.path;
    } else {
      console.log(await resp.json());
      console.log("fetched");
    }
  } catch (err) {
    console.error("Fetch failed:", err);
  }
}

// logout and and disable app area
async function logout() {
  try {
    const resp = await fetch("/api/logout", {
      method: "GET",
      credentials: "include",
    });

    if (resp.ok) {
      document.getElementById("appArea").style.display = "none";
      document.getElementById("loginArea").style.display = "block";
    } else {
    }
  } catch (err) {
    console.error("Fetch failed:", err);
  }
}

// fetch all downloads and also call filltable
async function getDownloadsAndFillTable() {
  try {
    const resp = await fetch("/api/downloads", {
      method: "GET",
      credentials: "include",
    });

    if (resp.ok) {
      const downloads = await resp.json();
      fillTable(downloads);
    } else {
      console.log(await resp.json());
    }
  } catch (err) {
    console.error("Fetch failed:", err);
  }
}

async function startDownload() {
  const data = {
    url: document.getElementById("url").value.trim(),
    dir: document.getElementById("dir").value.trim(),
    filename: document.getElementById("filename").value.trim(),
  };

  if (!data.url) {
    alert("URL is required");
    return;
  }

  try {
    const resp = await fetch("/api/add", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
      credentials: "include",
    });

    const respData = await resp.json();
    console.log(respData);
    if (resp.ok) {
      // write to UI that downloading has started
      document.getElementById("downloadInfo").textContent =
        "started downloading " + respData.filename;
    } else {
      document.getElementById("downloadInfo").textContent = respData.err;
    }
  } catch (err) {
    console.error("Fetch failed:", err);
  }
}

// check if cookie token is present / valid
async function checkSession() {
  try {
    const resp = await fetch("/api/login", {
      method: "POST",
      credentials: "include",
    });

    // enable app are and disable login
    if (resp.ok) {
      document.getElementById("appArea").style.display = "block";
      document.getElementById("loginArea").style.display = "none";
      const el = document.getElementById("appArea");

      runAfterLogin();
      // enable login and
    } else {
      document.getElementById("appArea").style.display = "none";
      document.getElementById("loginArea").style.display = "block";
    }
  } catch (err) {
    console.error("Fetch failed:", err);
  }
}

// run all this after successful login
function runAfterLogin() {
  getDirInfo();
  getDownloadsAndFillTable();
  // get data from server every 2 seconds
  setInterval(() => {
    getDownloadsAndFillTable();
  }, "2000");
}

async function login() {
  const data = {
    password: document.getElementById("password").value.trim(),
  };

  if (!data.password) {
    alert("password is required");
    return;
  }

  console.log(data);

  try {
    const resp = await fetch("/api/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
      credentials: "include", // REQUIRED for different domains
    });

    if (resp.ok) {
      console.log(await resp.json());
      document.getElementById("appArea").style.display = "block";
      document.getElementById("loginArea").style.display = "none";

      runAfterLogin();
    } else {
      document.getElementById("info").textContent = "wrong password";
    }
  } catch (err) {
    console.error("Fetch failed:", err);
  }
}

checkSession();
