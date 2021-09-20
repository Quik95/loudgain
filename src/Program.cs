using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Linq;
using System.Threading;
using System.Threading.Tasks;
using System.Threading.Tasks.Dataflow;
using ShellProgressBar;

namespace loudgain
{
    class Program
    {
        public static ProgressBarOptions ChildProgressBarOptions = new()
        {
            ProgressCharacter = '─',
            CollapseWhenFinished = true,
            DisableBottomPercentage = true,
        };

        public static ProgressBar MasterProgressBar;

        static async Task Main(string[] args)
        {
            var songs = new SongsList(args);

            var songsInAlbum = Album.GetSongsInAlbum(songs.Songs);
            Console.WriteLine(
                $"found {songsInAlbum.Keys.Count} albums in {songs.Songs.Count} songs");

            MasterProgressBar = new ProgressBar(songs.Songs.Count + songsInAlbum.Keys.Count, "Scanning songs",
                new ProgressBarOptions
                {
                    ProgressCharacter = '█',
                    CollapseWhenFinished = true,
                });


            var scanResults = new ConcurrentDictionary<string, ScanResult>(
                songs.Songs.Select(song => new KeyValuePair<string, ScanResult>(song, new ScanResult(song)))
            );

            var trackProgressBar =
                MasterProgressBar.Spawn(songs.Songs.Count, $"Scanning tracks: [0/{songs.Songs.Count}]",
                    ChildProgressBarOptions);
            var songsDone = 0;
            var scanTrack = new ActionBlock<string>(
                async song =>
                {
                    var res = await ScanResult.TrackScan(song);
                    scanResults[song].Track = res;
                    Interlocked.Increment(ref songsDone);

                    MasterProgressBar.Tick();
                    trackProgressBar.Tick($"Scanning tracks: [{songsDone}/{songs.Songs.Count}]");
                },
                new ExecutionDataflowBlockOptions {MaxDegreeOfParallelism = Environment.ProcessorCount});

            songs.Songs.ForEach(song => scanTrack.Post(song));
            scanTrack.Complete();


            var albumsThatNeedScanning = songsInAlbum.Where(entry => entry.Value.Length > 1)
                .Select(entry => entry.Value).ToArray();

            var albumProgressBar = MasterProgressBar.Spawn(songs.Songs.Count,
                $"Scanning albums: [0/{songsInAlbum.Keys.Count}]",
                ChildProgressBarOptions);
            var albumsDone = 0;
            var scanAlbum = new ActionBlock<string[]>(
                async albumSongs =>
                {
                    var result = await Album.ScanAlbum(albumSongs);
                    if (result is not null)
                    {
                        foreach (var song in albumSongs)
                        {
                            scanResults[song].Album = result;
                        }
                    }

                    Interlocked.Increment(ref albumsDone);

                    MasterProgressBar.Tick();
                    albumProgressBar.Tick($"Scanning albums: [{albumsDone}/{songsInAlbum.Keys.Count}]");
                }, new ExecutionDataflowBlockOptions {MaxDegreeOfParallelism = Environment.ProcessorCount});

            foreach (var albumSongs in albumsThatNeedScanning)
            {
                scanAlbum.Post(albumSongs);
            }

            scanAlbum.Complete();

            await Task.WhenAll(scanTrack.Completion, scanAlbum.Completion);

            trackProgressBar.Dispose();
            albumProgressBar.Dispose();
            MasterProgressBar.Dispose();

            foreach (var albumThatDoesNotNeedScanning in songsInAlbum.Where(entry => entry.Value.Length == 1)
                .Select(entry => entry.Value[0]))
            {
                if (scanResults[albumThatDoesNotNeedScanning].Track is not null)
                    scanResults[albumThatDoesNotNeedScanning].Album = scanResults[albumThatDoesNotNeedScanning].Track;
            }

            foreach (var scanResult in scanResults.Values)
            {
                Console.WriteLine(scanResult);
            }
        }
    }
}