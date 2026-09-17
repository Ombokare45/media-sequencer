import { useEffect, useState } from "react";
import "./App.css";
import MediaWindow from "./MediaWindow";

const API_URL =
  import.meta.env.VITE_API_URL || "http://localhost:8080";

function App() {
  const [windows, setWindows] = useState([]);
const [media, setMedia] = useState([]);
const [selectedWindow, setSelectedWindow] = useState("");
const [selectedMedia, setSelectedMedia] = useState("");
const [syncDuration, setSyncDuration] = useState(20);

const [syncState, setSyncState] = useState({
  active: false,
});

  // Load windows and playlists
  async function loadWindows() {
    try {
      const response = await fetch(`${API_URL}/api/windows`);
      const windowData = await response.json();

      const windowsWithPlaylists = await Promise.all(
        windowData.map(async (window) => {
          const playlistResponse = await fetch(
            `${API_URL}/api/windows/${window.id}/playlist`
          );

          const playlist = await playlistResponse.json();

          return {
            ...window,
            playlist,
          };
        })
      );

      setWindows(windowsWithPlaylists);

      if (!selectedWindow && windowsWithPlaylists.length > 0) {
        setSelectedWindow(String(windowsWithPlaylists[0].id));
      }
    } catch (error) {
      console.error("Failed to load windows:", error);
    }
  }

  // Load available media
  async function loadMedia() {
    try {
      const response = await fetch(`${API_URL}/api/media`);
      const data = await response.json();

      setMedia(data);

      if (!selectedMedia && data.length > 0) {
        setSelectedMedia(String(data[0].id));
      }
    } catch (error) {
      console.error("Failed to load media:", error);
    }
  }

 // Initial data loading + automatic playlist refresh
useEffect(() => {
  loadWindows();
  loadMedia();

  const interval = setInterval(() => {
    loadWindows();
  }, 5000);

  return () => clearInterval(interval);
}, []);

  // Check sync state every second
  useEffect(() => {
    async function loadSyncState() {
      try {
        const response = await fetch(`${API_URL}/api/sync`);
        const data = await response.json();

        setSyncState(data);
      } catch (error) {
        console.error("Failed to fetch sync state:", error);
      }
    }

    loadSyncState();

    const interval = setInterval(loadSyncState, 1000);

    return () => clearInterval(interval);
  }, []);

  // Add selected media to selected window
  async function addMediaToWindow() {
    if (!selectedWindow || !selectedMedia) {
      alert("Please select a window and media.");
      return;
    }

    try {
      const response = await fetch(
        `${API_URL}/api/windows/${selectedWindow}/playlist`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            media_id: Number(selectedMedia),
          }),
        }
      );

      if (!response.ok) {
        throw new Error("Failed to add media");
      }

      await response.json();

      // Reload playlists from backend
      await loadWindows();

      alert("Media added successfully.");
    } catch (error) {
      console.error("Failed to add media:", error);
      alert("Failed to add media.");
    }
  }
    // Start synchronized playback
  async function startSync() {
    if (!selectedMedia) {
      alert("Please select media.");
      return;
    }

    if (Number(syncDuration) <= 0) {
      alert("Sync duration must be greater than 0.");
      return;
    }

    try {
      const response = await fetch(`${API_URL}/api/sync`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          media_id: Number(selectedMedia),
          duration_seconds: Number(syncDuration),
        }),
      });

      if (!response.ok) {
        throw new Error("Failed to start sync");
      }

      const data = await response.json();

      setSyncState(data);
    } catch (error) {
      console.error("Failed to start sync:", error);
      alert("Failed to start synchronization.");
    }
  }
async function createMedia() {
  const name = document.getElementById("mediaName").value.trim();
  const type = document.getElementById("mediaType").value;
  const url = document.getElementById("mediaUrl").value.trim();
  const duration = Number(
    document.getElementById("mediaDuration").value
  );

  if (!name) {
    alert("Please enter media name.");
    return;
  }

  if (type !== "blank" && !url) {
    alert("Please enter media URL.");
    return;
  }

  if (duration <= 0) {
    alert("Duration must be greater than 0.");
    return;
  }

  try {
    const response = await fetch(`${API_URL}/api/media`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        name,
        type,
        url,
        duration_seconds: duration,
      }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText);
    }

    await response.json();

    await loadMedia();

    document.getElementById("mediaName").value = "";
    document.getElementById("mediaUrl").value = "";

    alert("Media created successfully.");
  } catch (error) {
    console.error("Failed to create media:", error);
    alert("Failed to create media.");
  }
}
  return (

    <div className="app">

      <header className="header">
        <h1>Media Sequencer</h1>

        <p>
          Multi-Window Synchronized Media Playback
        </p>

        {syncState.active && (
          <div className="sync-status">
            🔴 SYNC ACTIVE — {syncState.media_name}
          </div>
        )}
      </header>

      {/* Dynamic Playlist Controls */}
      <section className="control-panel">

        <h2>Playlist Management</h2>

        <div className="control-row">

          <label>
            Window:
            <select
              value={selectedWindow}
              onChange={(event) =>
                setSelectedWindow(event.target.value)
              }
            >
              {windows.map((window) => (
                <option
                  key={window.id}
                  value={window.id}
                >
                  {window.name}
                </option>
              ))}
            </select>
          </label>

          <label>
            Media:
            <select
              value={selectedMedia}
              onChange={(event) =>
                setSelectedMedia(event.target.value)
              }
            >
              {media.map((item) => (
                <option
                  key={item.id}
                  value={item.id}
                >
                  {item.name} ({item.type})
                </option>
              ))}
            </select>
          </label>

          <button onClick={addMediaToWindow}>
            ➕ Add Media
          </button>

        </div>

      </section>
      {/* Synchronization Controls */}
      <section className="control-panel sync-panel">

        <h2>Synchronization</h2>

        <div className="control-row">

          <label>
            Media:
            <select
              value={selectedMedia}
              onChange={(event) =>
                setSelectedMedia(event.target.value)
              }
            >
              {media.map((item) => (
                <option
                  key={item.id}
                  value={item.id}
                >
                  {item.name} ({item.type})
                </option>
              ))}
            </select>
          </label>

          <label>
            Duration (seconds):
            <input
              type="number"
              min="1"
              value={syncDuration}
              onChange={(event) =>
                setSyncDuration(event.target.value)
              }
            />
          </label>

          <button
            onClick={startSync}
            disabled={syncState.active}
          >
            🔴 {syncState.active ? "Sync Active" : "Start Sync"}
          </button>

        </div>

        {syncState.active && (
          <p className="sync-running">
            🔴 Synchronizing {syncState.media_name} for{" "}
            {syncState.duration_seconds} seconds
          </p>
        )}

      </section>
      {/* Create New Media */}
<section className="control-panel">

  <h2>Create New Media</h2>

  <div className="control-row">

    <label>
      Name:
      <input
        type="text"
        id="mediaName"
        placeholder="e.g. M7"
      />
    </label>

    <label>
      Type:
      <select id="mediaType">
        <option value="image">Image</option>
        <option value="video">Video</option>
        <option value="blank">Blank</option>
      </select>
    </label>

    <label>
      URL:
      <input
        type="text"
        id="mediaUrl"
        placeholder="https://..."
      />
    </label>

    <label>
      Duration (sec):
      <input
        type="number"
        id="mediaDuration"
        min="1"
        defaultValue="10"
      />
    </label>

    <button onClick={createMedia}>
      ➕ Create Media
    </button>

  </div>

</section>
      {/* Media Windows */}
      <main className="window-grid">
        {windows.map((window) => (
          <MediaWindow
            key={window.id}
            window={window}
            syncState={syncState}
          />
        ))}
      </main>

    </div>
  );
}

export default App;