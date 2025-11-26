#!/usr/bin/env python3
"""
YouTube Music API Client
Handles operations for liked songs and playlist management
"""

import sys
import json
from ytmusicapi import YTMusic


def get_liked_songs(oauth_file):
    """
    Get user's liked songs from YouTube Music

    Args:
        oauth_file: Path to OAuth credentials JSON file

    Returns:
        JSON string of liked songs
    """
    try:
        ytmusic = YTMusic(oauth_file)
        liked_songs = ytmusic.get_liked_songs(limit=None)

        # Extract relevant track information
        tracks = []
        for track in liked_songs.get('tracks', []):
            track_data = {
                'video_id': track.get('videoId', ''),
                'duration_ms': int(track.get('duration_seconds', 0)) * 1000,
                'title': track.get('title', ''),
                'channel_id': track.get('artists', [{}])[0].get('id', '') if track.get('artists') else '',
                'channel_name': track.get('artists', [{}])[0].get('name', '') if track.get('artists') else '',
                'thumbnail_url': track.get('thumbnails', [{}])[-1].get('url', '') if track.get('thumbnails') else '',
                'artists': [artist.get('name', '') for artist in track.get('artists', [])]
            }
            tracks.append(track_data)

        return json.dumps({'tracks': tracks, 'count': len(tracks)})
    except Exception as e:
        return json.dumps({'error': str(e)})


def create_playlist(oauth_file, title, description, video_ids):
    """
    Create a new playlist with specified tracks

    Args:
        oauth_file: Path to OAuth credentials JSON file
        title: Playlist title
        description: Playlist description
        video_ids: List of video IDs to add to playlist

    Returns:
        JSON string with playlist ID
    """
    try:
        ytmusic = YTMusic(oauth_file)

        # Create playlist
        playlist_id = ytmusic.create_playlist(title, description)

        # Add tracks to playlist if video_ids provided
        if video_ids and len(video_ids) > 0:
            ytmusic.add_playlist_items(playlist_id, video_ids)

        return json.dumps({'playlist_id': playlist_id, 'success': True})
    except Exception as e:
        return json.dumps({'error': str(e), 'success': False})


def main():
    """
    Main entry point for CLI execution
    Command format: python ytmusic_client.py <command> <oauth_file> [args...]
    """
    if len(sys.argv) < 3:
        print(json.dumps({'error': 'Usage: ytmusic_client.py <command> <oauth_file> [args...]'}))
        sys.exit(1)

    command = sys.argv[1]
    oauth_file = sys.argv[2]

    if command == 'get_liked_songs':
        result = get_liked_songs(oauth_file)
        print(result)

    elif command == 'create_playlist':
        if len(sys.argv) < 5:
            print(json.dumps({'error': 'Usage: ytmusic_client.py create_playlist <oauth_file> <title> <description> [video_ids_json]'}))
            sys.exit(1)

        title = sys.argv[3]
        description = sys.argv[4]
        video_ids = json.loads(sys.argv[5]) if len(sys.argv) > 5 else []

        result = create_playlist(oauth_file, title, description, video_ids)
        print(result)

    else:
        print(json.dumps({'error': f'Unknown command: {command}'}))
        sys.exit(1)


if __name__ == '__main__':
    main()
