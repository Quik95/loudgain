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
        static async Task Main(string[] args)
        {
            var songs = new SongsList(args);
            
            var songsInAlbum = Album.GetSongsInAlbum(songs.Songs);
            Console.WriteLine($"found {songsInAlbum.Keys.Count} albums in {songsInAlbum.Values.Aggregate(0, (cum, songs) => cum + songs.Length)} songs");

            var scanResults = new ConcurrentDictionary<string, ScanResult>(
                songs.Songs.Select(song => new KeyValuePair<string, ScanResult>(song, new ScanResult(song)))
            );
            var options = new ProgressBarOptions
            {
                ProgressCharacter = '█',
                CollapseWhenFinished = true,
                DisableBottomPercentage = true,
            };


            using (var pbar = new ProgressBar(songs.Songs.Count, $"Scanning tracks: [0/{songs.Songs.Count}]", options))
            {
                var songsDone = 0;
                var scanTrack = new ActionBlock<string>(
                    async song =>
                    {
                        var res = await ScanResult.TrackScan(song);
                        scanResults[song].Track = res;
                        Interlocked.Increment(ref songsDone);
                        pbar.Tick($"Scanning tracks: [{songsDone}/{songs.Songs.Count}]");
                    },
                    new ExecutionDataflowBlockOptions {MaxDegreeOfParallelism = Environment.ProcessorCount});

                songs.Songs.ForEach(song => scanTrack.Post(song));

                scanTrack.Complete();
                await scanTrack.Completion;
            }



            foreach (var songThatDoesNotNeedScanning in songsInAlbum.Where(entry => entry.Value.Length == 1).Select(entry => entry.Value[0]))
            {
                if (scanResults[songThatDoesNotNeedScanning].Track is not null)
                    scanResults[songThatDoesNotNeedScanning].Album = scanResults[songThatDoesNotNeedScanning].Track;
            }

            var albumsThatNeedScanning = songsInAlbum.Where(entry => entry.Value.Length > 1).Select(entry => entry.Value).ToArray();
            var numberOfAlbumsToScan = albumsThatNeedScanning.Length;
            using (var pbar = new ProgressBar(songs.Songs.Count, $"Scanning albums: [0/{numberOfAlbumsToScan}]",
                options))
            {
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
                        pbar.Tick($"Scanning albums: [{albumsDone}/{numberOfAlbumsToScan}]");
                    }, new ExecutionDataflowBlockOptions{MaxDegreeOfParallelism = Environment.ProcessorCount});
                
                foreach (var albumSongs in albumsThatNeedScanning)
                {
                    scanAlbum.Post(albumSongs);
                }
                
                scanAlbum.Complete();
                await scanAlbum.Completion;
            }

            foreach (var scanResult in scanResults.Values)
            {
                Console.WriteLine(scanResult);
            }
        }
    }
}