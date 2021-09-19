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

            var scanResults = new ConcurrentDictionary<string, ScanResult>(
                songs.Songs.Select(song => new KeyValuePair<string, ScanResult>(song, new ScanResult(song)))
            );
            var options = new ProgressBarOptions
            {
                ProgressCharacter = '█',
                CollapseWhenFinished = true,
                DisableBottomPercentage = true,
            };


            using var pbar = new ProgressBar(songs.Songs.Count, $"Scanning tracks: [0/{songs.Songs.Count}]", options);
            var songsDone = 0;
            var scanTrack = new ActionBlock<string>(
                async song =>
                {
                    var res = await ScanResult.TrackScan(song);
                    scanResults[song].Track = res;
                    Interlocked.Increment(ref songsDone);
                    pbar.Tick($"Scanning tracks: [{songsDone}/{songs.Songs.Count}]");
                },
                new ExecutionDataflowBlockOptions{MaxDegreeOfParallelism = Environment.ProcessorCount});

            songs.Songs.ForEach(song => scanTrack.Post(song));

            scanTrack.Complete();
            await scanTrack.Completion;

            foreach (var scanResult in scanResults.Values)
            {
                Console.WriteLine(scanResult);
            }
        }
    }
}