using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using ShellProgressBar;

namespace loudgain
{
    class Program
    {
        static async Task Main(string[] args)
        {
            var songs = new SongsList(args);

            var scanResults = new Dictionary<string, ScanResult>(
                songs.Songs.Select(song => new KeyValuePair<string, ScanResult>(song, new ScanResult(song)))
            );
            var options = new ProgressBarOptions
            {
                ProgressCharacter = '█',
                CollapseWhenFinished = true,
            };

            using (var pbar = new ProgressBar(songs.Songs.Count, "Scanning tracks...", options))
            {
                foreach (var song in songs.Songs)
                {
                    var res = await ScanResult.TrackScan(song);
                    scanResults[song].Track = res;
                    pbar.Tick();
                }
            }

            foreach (var keyValuePair in scanResults)
            {
                Console.WriteLine(keyValuePair.Value);
            }
        }
    }
}