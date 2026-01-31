package spotify

import (
	"context"
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
	return c.client.PlayerCurrentlyPlaying(ctx)
}

func (c *Client) PlayerState(ctx context.Context) (*spotify.PlayerState, error) {
	return c.client.PlayerState(ctx)
}

func (c *Client) Play(ctx context.Context) error {
	return c.client.Play(ctx)
}

func (c *Client) Pause(ctx context.Context) error {
	return c.client.Pause(ctx)
}

func (c *Client) Next(ctx context.Context) error {
	return c.client.Next(ctx)
}

func (c *Client) SkipToNth(ctx context.Context, n int) error {
	for i := 0; i < n; i++ {
		if err := c.client.Next(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Previous(ctx context.Context) error {
	return c.client.Previous(ctx)
}

func (c *Client) Seek(ctx context.Context, position time.Duration) error {
	return c.client.Seek(ctx, int(position.Milliseconds()))
}

func (c *Client) UserPlaylists(ctx context.Context) ([]spotify.SimplePlaylist, error) {
	playlists, err := c.client.CurrentUsersPlaylists(ctx)
	if err != nil {
		return nil, err
	}
	return playlists.Playlists, nil
}

func (c *Client) PlaylistTracks(ctx context.Context, playlistID spotify.ID) ([]spotify.PlaylistTrack, error) {
	tracks, err := c.client.GetPlaylistTracks(ctx, playlistID)
	if err != nil {
		return nil, err
	}
	return tracks.Tracks, nil
}

func (c *Client) SavedTracks(ctx context.Context) ([]spotify.SavedTrack, error) {
	// Liked Songsを取得（最大50件ずつ）
	var allTracks []spotify.SavedTrack
	limit := 50
	offset := 0

	for {
		tracks, err := c.client.CurrentUsersTracks(ctx, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			return nil, err
		}

		allTracks = append(allTracks, tracks.Tracks...)

		if len(tracks.Tracks) < limit {
			break
		}
		offset += limit
	}

	return allTracks, nil
}

func (c *Client) PlayTrackInContext(ctx context.Context, contextURI spotify.URI, offset int) error {
	opts := &spotify.PlayOptions{
		PlaybackContext: &contextURI,
		PlaybackOffset:  &spotify.PlaybackOffset{Position: &offset},
	}
	return c.client.PlayOpt(ctx, opts)
}

func (c *Client) PlayTrackFromURIList(ctx context.Context, uris []spotify.URI, offset int) error {
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
	return c.client.PlayOpt(ctx, opts)
}

func (c *Client) PlayLikedSongs(ctx context.Context, userID string, offset int) error {
	// Liked Songs collection URI: spotify:user:<user_id>:collection
	collectionURI := spotify.URI("spotify:user:" + userID + ":collection")
	opts := &spotify.PlayOptions{
		PlaybackContext: &collectionURI,
		PlaybackOffset:  &spotify.PlaybackOffset{Position: &offset},
	}
	return c.client.PlayOpt(ctx, opts)
}

func (c *Client) PlayTrackAlone(ctx context.Context, trackURI spotify.URI) error {
	opts := &spotify.PlayOptions{
		URIs: []spotify.URI{trackURI},
	}
	return c.client.PlayOpt(ctx, opts)
}

func (c *Client) ToggleShuffle(ctx context.Context, shuffle bool) error {
	return c.client.Shuffle(ctx, shuffle)
}

func (c *Client) SetRepeat(ctx context.Context, state string) error {
	return c.client.Repeat(ctx, state)
}

func (c *Client) Search(ctx context.Context, query string) ([]spotify.FullTrack, error) {
	results, err := c.client.Search(ctx, query, spotify.SearchTypeTrack)
	if err != nil {
		return nil, err
	}

	if results.Tracks == nil {
		return []spotify.FullTrack{}, nil
	}

	return results.Tracks.Tracks, nil
}

func (c *Client) CurrentUser(ctx context.Context) (*spotify.PrivateUser, error) {
	return c.client.CurrentUser(ctx)
}

func (c *Client) GetQueue(ctx context.Context) (*spotify.Queue, error) {
	return c.client.GetQueue(ctx)
}

func (c *Client) PlayerDevices(ctx context.Context) ([]spotify.PlayerDevice, error) {
	return c.client.PlayerDevices(ctx)
}

func (c *Client) SetVolume(ctx context.Context, volume int) error {
	return c.client.Volume(ctx, volume)
}

func (c *Client) TransferPlayback(ctx context.Context, deviceID spotify.ID) error {
	return c.client.TransferPlayback(ctx, deviceID, true)
}
