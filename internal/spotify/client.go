package spotify

import (
	"context"
	"spotify-tui/internal/logger"
	"time"

	"github.com/zmb3/spotify/v2"
)

type Client struct {
	client *spotify.Client
}

func NewClient(c *spotify.Client) *Client {
	return &Client{client: c}
}

func (c *Client) CurrentlyPlaying(ctx context.Context) (*spotify.CurrentlyPlaying, error) {
	logger.Debug("API call", "method", "CurrentlyPlaying")
	result, err := c.client.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		logger.Error("API error", "method", "CurrentlyPlaying", "error", err)
	}
	return result, err
}

func (c *Client) PlayerState(ctx context.Context) (*spotify.PlayerState, error) {
	logger.Debug("API call", "method", "PlayerState")
	result, err := c.client.PlayerState(ctx)
	if err != nil {
		logger.Error("API error", "method", "PlayerState", "error", err)
	}
	return result, err
}

func (c *Client) Play(ctx context.Context) error {
	logger.Debug("API call", "method", "Play")
	err := c.client.Play(ctx)
	if err != nil {
		logger.Error("API error", "method", "Play", "error", err)
	}
	return err
}

func (c *Client) Pause(ctx context.Context) error {
	logger.Debug("API call", "method", "Pause")
	err := c.client.Pause(ctx)
	if err != nil {
		logger.Error("API error", "method", "Pause", "error", err)
	}
	return err
}

func (c *Client) Next(ctx context.Context) error {
	logger.Debug("API call", "method", "Next")
	err := c.client.Next(ctx)
	if err != nil {
		logger.Error("API error", "method", "Next", "error", err)
	}
	return err
}

func (c *Client) SkipToNth(ctx context.Context, n int) error {
	logger.Debug("API call", "method", "SkipToNth", "n", n)
	for i := 0; i < n; i++ {
		if err := c.client.Next(ctx); err != nil {
			logger.Error("API error", "method", "SkipToNth", "error", err)
			return err
		}
	}
	return nil
}

func (c *Client) Previous(ctx context.Context) error {
	logger.Debug("API call", "method", "Previous")
	err := c.client.Previous(ctx)
	if err != nil {
		logger.Error("API error", "method", "Previous", "error", err)
	}
	return err
}

func (c *Client) Seek(ctx context.Context, position time.Duration) error {
	logger.Debug("API call", "method", "Seek", "position", position)
	err := c.client.Seek(ctx, int(position.Milliseconds()))
	if err != nil {
		logger.Error("API error", "method", "Seek", "error", err)
	}
	return err
}

func (c *Client) UserPlaylists(ctx context.Context) ([]spotify.SimplePlaylist, error) {
	logger.Debug("API call", "method", "UserPlaylists")
	playlists, err := c.client.CurrentUsersPlaylists(ctx)
	if err != nil {
		logger.Error("API error", "method", "UserPlaylists", "error", err)
		return nil, err
	}
	return playlists.Playlists, nil
}

func (c *Client) PlaylistTracks(ctx context.Context, playlistID spotify.ID) ([]spotify.PlaylistTrack, error) {
	logger.Debug("API call", "method", "PlaylistTracks", "playlistID", playlistID)
	tracks, err := c.client.GetPlaylistTracks(ctx, playlistID)
	if err != nil {
		logger.Error("API error", "method", "PlaylistTracks", "error", err)
		return nil, err
	}
	return tracks.Tracks, nil
}

func (c *Client) SavedTracks(ctx context.Context) ([]spotify.SavedTrack, error) {
	logger.Debug("API call", "method", "SavedTracks")
	// Liked Songsを取得（最大50件ずつ）
	var allTracks []spotify.SavedTrack
	limit := 50
	offset := 0

	for {
		tracks, err := c.client.CurrentUsersTracks(ctx, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			logger.Error("API error", "method", "SavedTracks", "error", err)
			return nil, err
		}

		allTracks = append(allTracks, tracks.Tracks...)

		if len(tracks.Tracks) < limit {
			break
		}
		offset += limit
	}

	logger.Debug("API call completed", "method", "SavedTracks", "totalTracks", len(allTracks))
	return allTracks, nil
}

func (c *Client) PlayTrackInContext(ctx context.Context, contextURI spotify.URI, offset int) error {
	logger.Debug("API call", "method", "PlayTrackInContext", "contextURI", contextURI, "offset", offset)
	opts := &spotify.PlayOptions{
		PlaybackContext: &contextURI,
		PlaybackOffset:  &spotify.PlaybackOffset{Position: &offset},
	}
	err := c.client.PlayOpt(ctx, opts)
	if err != nil {
		logger.Error("API error", "method", "PlayTrackInContext", "error", err)
	}
	return err
}

func (c *Client) PlayTrackFromURIList(ctx context.Context, uris []spotify.URI, offset int) error {
	logger.Debug("API call", "method", "PlayTrackFromURIList", "uriCount", len(uris), "offset", offset)
	if len(uris) == 0 {
		return nil
	}

	if offset >= len(uris) {
		offset = 0
	}

	opts := &spotify.PlayOptions{
		URIs:           uris,
		PlaybackOffset: &spotify.PlaybackOffset{Position: &offset},
	}
	err := c.client.PlayOpt(ctx, opts)
	if err != nil {
		logger.Error("API error", "method", "PlayTrackFromURIList", "error", err)
	}
	return err
}

func (c *Client) PlayLikedSongs(ctx context.Context, userID string, offset int) error {
	logger.Debug("API call", "method", "PlayLikedSongs", "userID", userID, "offset", offset)
	// Liked Songs collection URI: spotify:user:<user_id>:collection
	collectionURI := spotify.URI("spotify:user:" + userID + ":collection")
	opts := &spotify.PlayOptions{
		PlaybackContext: &collectionURI,
		PlaybackOffset:  &spotify.PlaybackOffset{Position: &offset},
	}
	err := c.client.PlayOpt(ctx, opts)
	if err != nil {
		logger.Error("API error", "method", "PlayLikedSongs", "error", err)
	}
	return err
}

func (c *Client) PlayTrackAlone(ctx context.Context, trackURI spotify.URI) error {
	logger.Debug("API call", "method", "PlayTrackAlone", "trackURI", trackURI)
	opts := &spotify.PlayOptions{
		URIs: []spotify.URI{trackURI},
	}
	err := c.client.PlayOpt(ctx, opts)
	if err != nil {
		logger.Error("API error", "method", "PlayTrackAlone", "error", err)
	}
	return err
}

func (c *Client) ToggleShuffle(ctx context.Context, shuffle bool) error {
	logger.Debug("API call", "method", "ToggleShuffle", "shuffle", shuffle)
	err := c.client.Shuffle(ctx, shuffle)
	if err != nil {
		logger.Error("API error", "method", "ToggleShuffle", "error", err)
	}
	return err
}

func (c *Client) SetRepeat(ctx context.Context, state string) error {
	logger.Debug("API call", "method", "SetRepeat", "state", state)
	err := c.client.Repeat(ctx, state)
	if err != nil {
		logger.Error("API error", "method", "SetRepeat", "error", err)
	}
	return err
}

func (c *Client) Search(ctx context.Context, query string) ([]spotify.FullTrack, error) {
	logger.Debug("API call", "method", "Search", "query", query)
	results, err := c.client.Search(ctx, query, spotify.SearchTypeTrack)
	if err != nil {
		logger.Error("API error", "method", "Search", "error", err)
		return nil, err
	}

	if results.Tracks == nil {
		return []spotify.FullTrack{}, nil
	}

	logger.Debug("API call completed", "method", "Search", "resultCount", len(results.Tracks.Tracks))
	return results.Tracks.Tracks, nil
}

func (c *Client) CurrentUser(ctx context.Context) (*spotify.PrivateUser, error) {
	logger.Debug("API call", "method", "CurrentUser")
	result, err := c.client.CurrentUser(ctx)
	if err != nil {
		logger.Error("API error", "method", "CurrentUser", "error", err)
	}
	return result, err
}

func (c *Client) GetQueue(ctx context.Context) (*spotify.Queue, error) {
	logger.Debug("API call", "method", "GetQueue")
	result, err := c.client.GetQueue(ctx)
	if err != nil {
		logger.Error("API error", "method", "GetQueue", "error", err)
	}
	return result, err
}

func (c *Client) PlayerDevices(ctx context.Context) ([]spotify.PlayerDevice, error) {
	logger.Debug("API call", "method", "PlayerDevices")
	result, err := c.client.PlayerDevices(ctx)
	if err != nil {
		logger.Error("API error", "method", "PlayerDevices", "error", err)
	}
	return result, err
}

func (c *Client) SetVolume(ctx context.Context, volume int) error {
	logger.Debug("API call", "method", "SetVolume", "volume", volume)
	err := c.client.Volume(ctx, volume)
	if err != nil {
		logger.Error("API error", "method", "SetVolume", "error", err)
	}
	return err
}

func (c *Client) TransferPlayback(ctx context.Context, deviceID spotify.ID) error {
	logger.Debug("API call", "method", "TransferPlayback", "deviceID", deviceID)
	err := c.client.TransferPlayback(ctx, deviceID, true)
	if err != nil {
		logger.Error("API error", "method", "TransferPlayback", "error", err)
	}
	return err
}
