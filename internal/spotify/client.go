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

func (c *Client) Play(ctx context.Context) error {
	return c.client.Play(ctx)
}

func (c *Client) Pause(ctx context.Context) error {
	return c.client.Pause(ctx)
}

func (c *Client) Next(ctx context.Context) error {
	return c.client.Next(ctx)
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
