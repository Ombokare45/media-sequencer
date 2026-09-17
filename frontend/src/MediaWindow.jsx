import { useEffect, useRef, useState } from "react";

const FIVE_HOURS = 5 * 60 * 60 * 1000;

function MediaWindow({ window, syncState }) {
  const [currentIndex, setCurrentIndex] = useState(0);
const [mediaError, setMediaError] = useState(false);

const syncVideoRef = useRef(null);
const cycleStartRef = useRef(Date.now());

const playlist = window.playlist || [];
const currentItem = playlist[currentIndex];

useEffect(() => {
  setMediaError(false);
}, [currentItem?.id]);

  // --------------------------------------------------
  // 5-HOUR NORMAL PLAYBACK CYCLE
  // --------------------------------------------------

  useEffect(() => {
    if (playlist.length === 0 || syncState?.active) {
      return;
    }

    const updateCurrentMedia = () => {
      const now = Date.now();

      // Position inside the current 5-hour cycle
      const cycleElapsed =
        (now - cycleStartRef.current) % FIVE_HOURS;

     const playlistTotalMs = playlist.reduce(
  (total, item) =>
    total +
    Math.max(item.duration_seconds, 1) * 1000,
  0
);

// Repeat the playlist continuously
const positionInPlaylist =
  cycleElapsed % playlistTotalMs;

let elapsed = positionInPlaylist;
let selectedIndex = 0;

for (let i = 0; i < playlist.length; i++) {
  const duration =
    Math.max(playlist[i].duration_seconds, 1) * 1000;

  if (elapsed < duration) {
    selectedIndex = i;
    break;
  }

  elapsed -= duration;
}

      setCurrentIndex(selectedIndex);
    };

    updateCurrentMedia();

    const interval = setInterval(
      updateCurrentMedia,
      500
    );

    return () => clearInterval(interval);
  }, [
    playlist,
    syncState?.active,
  ]);

  // --------------------------------------------------
  // SYNCHRONIZED VIDEO POSITION
  // --------------------------------------------------

  useEffect(() => {
    if (
      !syncState?.active ||
      syncState.media_type !== "video" ||
      !syncState.started_at ||
      !syncVideoRef.current
    ) {
      return;
    }

    const startTime =
      new Date(syncState.started_at).getTime();

    if (Number.isNaN(startTime)) {
      return;
    }

    const elapsedSeconds =
      (Date.now() - startTime) / 1000;

    const mediaDuration =
      syncState.duration_seconds || 1;

    const position =
      elapsedSeconds % mediaDuration;

    syncVideoRef.current.currentTime = position;
  }, [
    syncState?.active,
    syncState?.started_at,
    syncState?.media_type,
    syncState?.duration_seconds,
  ]);

  // --------------------------------------------------
  // SYNC PLAYBACK
  // --------------------------------------------------

  if (syncState?.active) {
    return (
      <div className="media-window">

        <div className="window-header">
          <h2>{window.name}</h2>

          <span>
            🔴 SYNC: {syncState.media_name}
          </span>
        </div>

        <div className="display-area">

          {syncState.media_type === "image" && (
            <img
              src={syncState.url}
              alt={syncState.media_name}
              className="media-content"
            />
          )}

          {syncState.media_type === "video" && (
            <video
  key={syncState.started_at}
  ref={syncVideoRef}
  src={syncState.url}
  className="media-content"
  autoPlay
  muted
  playsInline
  onLoadedMetadata={(event) => {
    const startTime = new Date(syncState.started_at).getTime();

    if (Number.isNaN(startTime)) {
      return;
    }

    const elapsedSeconds =
      (Date.now() - startTime) / 1000;

    const position =
      elapsedSeconds % Math.max(syncState.duration_seconds, 1);

    event.currentTarget.currentTime = position;

    event.currentTarget
      .play()
      .catch((error) => {
        console.log("Video autoplay prevented:", error);
      });
  }}
/>
          )}

          {syncState.media_type === "blank" && (
            <div className="blank-screen"></div>
          )}

          <div className="media-info">
            🔴 Synchronized: {syncState.media_name}
          </div>

        </div>

      </div>
    );
  }

  // --------------------------------------------------
  // EMPTY PLAYLIST
  // --------------------------------------------------

  if (playlist.length === 0) {
    return (
      <div className="media-window">

        <div className="window-header">
          <h2>{window.name}</h2>

          <span>
            Window ID: {window.id}
          </span>
        </div>

        <div className="display-area">
          <p>No media configured</p>
        </div>

      </div>
    );
  }

  // --------------------------------------------------
  // NORMAL PLAYLIST DISPLAY
  // --------------------------------------------------

  return (
    <div className="media-window">

      <div className="window-header">

        <h2>{window.name}</h2>

        <span>
          Playing: {currentItem.media_name}
        </span>

      </div>

      <div className="display-area">

        {currentItem.media_type === "image" && !mediaError && (
  <img
    src={currentItem.url}
    alt={currentItem.media_name}
    className="media-content"
    onError={() => setMediaError(true)}
  />
)}

{currentItem.media_type === "image" && mediaError && (
  <div className="fallback-screen">
    <strong>Media unavailable</strong>
  </div>
)}

        {currentItem.media_type === "video" && !mediaError && (
  <video
    src={currentItem.url}
    className="media-content"
    autoPlay
    muted
    playsInline
    onError={() => setMediaError(true)}
  />
)}

{currentItem.media_type === "video" && mediaError && (
  <div className="fallback-screen">
    <strong>Media unavailable</strong>
  </div>
)}

        {currentItem.media_type === "blank" && (
          <div className="blank-screen"></div>
        )}
        {!["image", "video", "blank"].includes(currentItem.media_type) && (
  <div className="fallback-screen">
    <strong>Unsupported media type</strong>
  </div>
)}

        <div className="media-info">
          {currentItem.media_name} •{" "}
          {currentItem.duration_seconds}s
        </div>

      </div>

    </div>
  );
}

export default MediaWindow;