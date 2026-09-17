import { useEffect, useMemo, useRef, useState } from "react";

const CYCLE_SECONDS = 5 * 60 * 60; // 5 hours

function MediaWindow({ window, syncState }) {
  const normalVideoRef = useRef(null);
  const syncVideoRef = useRef(null);
  const cycleStartRef = useRef(Date.now());

  const [currentIndex, setCurrentIndex] = useState(0);
  const [mediaError, setMediaError] = useState(false);

  const playlist = window.playlist || [];

  /*
   * Normalize playlist media.
   *
   * Playlist API uses:
   *   media_type
   *
   * Sync API uses:
   *   type
   *
   * We convert both into the same "type" field.
   */
  const normalizedPlaylist = useMemo(() => {
    return playlist.map((item) => ({
      ...item,
      type: item.type || item.media_type || "",
      name: item.media_name || item.name || `Media ${item.media_id}`,
    }));
  }, [playlist]);

  /*
   * Reset playback timing only when the playlist
   * actually becomes empty/non-empty.
   */
  useEffect(() => {
    if (playlist.length === 0) {
      setCurrentIndex(0);
      cycleStartRef.current = Date.now();
    }
  }, [playlist.length]);

  /*
   * Determine the normal playlist item based on elapsed
   * time inside the 5-hour cycle.
   */
  useEffect(() => {
    if (normalizedPlaylist.length === 0) {
      return;
    }

    const calculateCurrentItem = () => {
      const elapsed =
        (Date.now() - cycleStartRef.current) / 1000;

      const cyclePosition =
        elapsed % CYCLE_SECONDS;

      let accumulated = 0;
      let selectedIndex = 0;

      /*
       * The playlist repeats continuously.
       * We use the configured duration of each media item.
       */
      const totalPlaylistDuration =
        normalizedPlaylist.reduce(
          (total, item) =>
            total + Math.max(Number(item.duration_seconds) || 1, 1),
          0
        );

      if (totalPlaylistDuration > 0) {
        const playlistPosition =
          cyclePosition % totalPlaylistDuration;

        accumulated = 0;

        for (let i = 0; i < normalizedPlaylist.length; i++) {
          const duration = Math.max(
            Number(normalizedPlaylist[i].duration_seconds) || 1,
            1
          );

          accumulated += duration;

          if (playlistPosition < accumulated) {
            selectedIndex = i;
            break;
          }
        }
      }

      setCurrentIndex(selectedIndex);
    };

    calculateCurrentItem();

    const interval = setInterval(
      calculateCurrentItem,
      500
    );

    return () => clearInterval(interval);
  }, [normalizedPlaylist]);

  /*
   * Clear media error whenever the displayed item changes.
   */
  useEffect(() => {
    setMediaError(false);
  }, [currentIndex, syncState?.active, syncState?.media_id]);

  const currentItem =
    normalizedPlaylist.length > 0
      ? normalizedPlaylist[currentIndex] ||
        normalizedPlaylist[0]
      : null;

  /*
   * Synchronization media.
   *
   * Sync API returns:
   * {
   *   media_id,
   *   media_name,
   *   type,
   *   url,
   *   duration_seconds,
   *   started_at
   * }
   */
  const syncItem = syncState?.active
    ? {
        id: syncState.media_id,
        media_id: syncState.media_id,
        name: syncState.media_name,
        type: syncState.type || syncState.media_type || "",
        url: syncState.url || "",
        duration_seconds:
          Number(syncState.duration_seconds) || 1,
      }
    : null;

  /*
   * Keep synchronized video position aligned using the
   * backend's started_at timestamp.
   */
  useEffect(() => {
    if (
      !syncState?.active ||
      !syncVideoRef.current ||
      syncItem?.type !== "video" ||
      !syncState.started_at
    ) {
      return;
    }

    const video = syncVideoRef.current;

    const synchronizeVideo = () => {
      const startedAt = new Date(
        syncState.started_at
      ).getTime();

      const elapsed =
        (Date.now() - startedAt) / 1000;

      if (elapsed < 0) {
        return;
      }

      /*
       * Use the actual video duration when available.
       * This prevents seeking beyond the real video length.
       */
      const actualDuration =
        Number.isFinite(video.duration) &&
        video.duration > 0
          ? video.duration
          : Number(syncState.duration_seconds) || 1;

      const targetTime =
        elapsed % actualDuration;

      /*
       * Only seek when the difference is meaningful.
       */
      if (
        Number.isFinite(video.currentTime) &&
        Math.abs(video.currentTime - targetTime) > 0.5
      ) {
        try {
          video.currentTime = targetTime;
        } catch {
          // Browser may reject seeking before metadata loads.
        }
      }

      video
        .play()
        .catch(() => {
          // Muted autoplay should normally succeed.
        });
    };

    synchronizeVideo();

    const interval = setInterval(
      synchronizeVideo,
      500
    );

    return () => clearInterval(interval);
  }, [
    syncState?.active,
    syncState?.started_at,
    syncState?.media_id,
    syncItem?.type,
  ]);

  /*
   * Keep normal videos playing.
   */
  useEffect(() => {
    if (
      syncState?.active ||
      !currentItem ||
      currentItem.type !== "video" ||
      !normalVideoRef.current
    ) {
      return;
    }

    const video = normalVideoRef.current;

    video
      .play()
      .catch(() => {
        // Muted autoplay should normally work.
      });
  }, [
    currentItem?.id,
    currentItem?.url,
    currentItem?.type,
    syncState?.active,
  ]);

  function renderMedia(item, mode = "normal") {
    if (!item) {
      return (
        <div className="media-placeholder">
          No media configured
        </div>
      );
    }

    /*
     * IMPORTANT:
     * Support both "type" and "media_type".
     */
    const type =
      item.type || item.media_type || "";

    const name =
      item.name ||
      item.media_name ||
      `Media ${item.media_id}`;

    const url = item.url || "";

    /*
     * IMAGE
     */
    if (type === "image") {
      return (
        <img
          key={`${mode}-${item.id}-${url}`}
          src={url}
          alt={name}
          className="media-content"
          onError={() => setMediaError(true)}
        />
      );
    }

    /*
     * VIDEO
     */
    if (type === "video") {
      if (!url) {
        return (
          <div className="media-placeholder">
            Video URL missing: {name}
          </div>
        );
      }

      if (mode === "sync") {
        return (
          <video
            key={`sync-${item.id}-${syncState?.started_at}`}
            ref={syncVideoRef}
            src={url}
            className="media-content"
            autoPlay
            muted
            playsInline
            preload="auto"
            onLoadedMetadata={(event) => {
              const video = event.currentTarget;

              if (!syncState?.started_at) {
                return;
              }

              const startedAt = new Date(
                syncState.started_at
              ).getTime();

              const elapsed =
                (Date.now() - startedAt) / 1000;

              if (elapsed >= 0) {
                const duration =
                  Number.isFinite(video.duration) &&
                  video.duration > 0
                    ? video.duration
                    : Number(
                        syncState.duration_seconds
                      ) || 1;

                video.currentTime =
                  elapsed % duration;
              }

              video.play().catch(() => {});
            }}
            onCanPlay={(event) => {
              event.currentTarget
                .play()
                .catch(() => {});
            }}
            onError={() => setMediaError(true)}
          />
        );
      }

      return (
        <video
          key={`normal-${item.id}-${url}`}
          ref={normalVideoRef}
          src={url}
          className="media-content"
          autoPlay
          muted
          playsInline
          preload="auto"
          loop
          onCanPlay={(event) => {
            event.currentTarget
              .play()
              .catch(() => {});
          }}
          onError={() => setMediaError(true)}
        />
      );
    }

    /*
     * BLANK
     */
    if (type === "blank") {
      return (
        <div
          key={`${mode}-blank-${item.id}`}
          className="media-content blank-media"
        />
      );
    }

    /*
     * Unknown media type
     */
    return (
      <div className="media-placeholder">
        Unsupported media type: {type || "unknown"} — {name}
      </div>
    );
  }

  /*
   * SYNC MODE
   */
  if (syncState?.active && syncItem) {
    return (
      <section className="media-window">
        <header className="window-header">
          <h2>{window.name}</h2>

          <span className="sync-badge">
            🔴 SYNC: {syncItem.name}
          </span>
        </header>

        <div className="media-area">
          {mediaError ? (
            <div className="media-placeholder">
              Unable to play {syncItem.name}
              <br />
              <small>{syncItem.url}</small>
            </div>
          ) : (
            renderMedia(syncItem, "sync")
          )}
        </div>

        <footer className="window-footer">
          🔴 Synchronized:{" "}
          <strong>{syncItem.name}</strong>
        </footer>
      </section>
    );
  }

  /*
   * NORMAL MODE
   */
  return (
    <section className="media-window">
      <header className="window-header">
        <h2>{window.name}</h2>

        {currentItem && (
          <span className="normal-badge">
            ▶ {currentItem.name}
          </span>
        )}
      </header>

      <div className="media-area">
        {mediaError ? (
          <div className="media-placeholder">
            Unable to play{" "}
            {currentItem?.name || "media"}
            <br />
            <small>{currentItem?.url}</small>
          </div>
        ) : (
          renderMedia(currentItem, "normal")
        )}
      </div>

      <footer className="window-footer">
        {currentItem ? (
          <>
            ▶ Playing:{" "}
            <strong>{currentItem.name}</strong>
          </>
        ) : (
          "No playlist configured"
        )}
      </footer>
    </section>
  );
}

export default MediaWindow;