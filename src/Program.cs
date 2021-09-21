using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Linq;
using System.Threading;
using System.Threading.Tasks;
using System.Threading.Tasks.Dataflow;
using CommandLine;
using ShellProgressBar;

namespace loudgain
{
    internal static class Program
    {
        private static readonly ProgressBarOptions ChildProgressBarOptions = new()
        {
            ProgressCharacter = '─',
            CollapseWhenFinished = true,
            DisableBottomPercentage = true,
        };

        public static ProgressBar MasterProgressBar = null!;
        private static ChildProgressBar _trackProgressBar = null!;
        private static ChildProgressBar _albumProgressBar = null!;

        public static Options Config = null!;

        private static readonly ConcurrentDictionary<string, ScanResult> ScanResults = new();

        static async Task Main(string[] args)
        {
            _parseArguments(args);
            Console.WriteLine(Config);
            
            var songs = new SongsList(Config.Songs);

            var songsInAlbum = Album.GetSongsInAlbum(songs.Songs);
            if (songsInAlbum.Keys.Count > 0)
                Console.WriteLine(
                    $"found {songsInAlbum.Keys.Count} albums in {songs.Songs.Count} songs");

            MasterProgressBar = new ProgressBar(songs.Songs.Count + songsInAlbum.Keys.Count, "Scanning songs",
                new ProgressBarOptions
                {
                    ProgressCharacter = '█',
                    CollapseWhenFinished = true,
                });


            foreach (var keyValuePair in songs.Songs.Select(song =>
                new KeyValuePair<string, ScanResult>(song, new ScanResult(song))))
            {
                ScanResults.TryAdd(keyValuePair.Key, keyValuePair.Value);
            }


            _trackProgressBar =
                MasterProgressBar.Spawn(songs.Songs.Count, $"Scanning tracks: [0/{songs.Songs.Count}]",
                    ChildProgressBarOptions);
            var trackAction = GetScanTrackAction(songs.Songs.Count);
            songs.Songs.ForEach(song => trackAction.Post(song));
            trackAction.Complete();


            // scan only albums with more than one track
            var albumsThatNeedScanning = songsInAlbum.Where(entry => entry.Value.Length > 1)
                .Select(entry => entry.Value).ToArray();

            _albumProgressBar = MasterProgressBar.Spawn(songs.Songs.Count,
                $"Scanning albums: [0/{songsInAlbum.Keys.Count}]",
                ChildProgressBarOptions);
            var albumAction = GetScanAlbumAction(songsInAlbum.Keys.Count);

            foreach (var albumSongs in albumsThatNeedScanning)
            {
                albumAction.Post(albumSongs);
            }

            albumAction.Complete();

            await Task.WhenAll(trackAction.Completion, albumAction.Completion);

            _trackProgressBar.Dispose();
            _albumProgressBar.Dispose();
            MasterProgressBar.Dispose();

            // if an albums has only one track, an album gain is equal to a track gain
            SetAlbumGainToTrackGain(songsInAlbum);

            foreach (var scanResult in ScanResults.Values)
            {
                Console.WriteLine(scanResult);
            }
        }

        private static void _parseArguments(string[] args)
        {
            var config = Parser.Default.ParseArguments<Options>(args).WithNotParsed(errors =>
            {
                if (errors.Any(error => error is CommandLine.HelpRequestedError))
                    Environment.Exit(0);

                foreach (var error in errors)
                {
                    Console.Error.WriteLine(error);
                }

                Environment.Exit(1);
            }).WithParsed(opt =>
            {
                opt.Pregain = new Decibel(opt._pregainFloat);
                opt.MaxTruePeakLevel = new Decibel(opt._maxTruePeakLevelFloat);
                Config = opt;
            });
        }

        private static ActionBlock<string> GetScanTrackAction(int songsCount
        )
        {
            var songsDone = 0;

            var scanTrack = new ActionBlock<string>(
                async song =>
                {
                    var res = await ScanResult.TrackScan(song);
                    ScanResults[song].Track = res;
                    Interlocked.Increment(ref songsDone);

                    MasterProgressBar.Tick();
                    _trackProgressBar.Tick($"Scanning tracks: [{songsDone}/{songsCount}]");
                },
                new ExecutionDataflowBlockOptions {MaxDegreeOfParallelism = Environment.ProcessorCount});

            return scanTrack;
        }

        private static ActionBlock<string[]> GetScanAlbumAction(int albumCount)
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
                            ScanResults[song].Album = result;
                        }
                    }

                    Interlocked.Increment(ref albumsDone);

                    MasterProgressBar.Tick();
                    _albumProgressBar.Tick($"Scanning albums: [{albumsDone}/{albumCount}]");
                }, new ExecutionDataflowBlockOptions {MaxDegreeOfParallelism = Environment.ProcessorCount});

            return scanAlbum;
        }

        private static void SetAlbumGainToTrackGain(Dictionary<string, string[]> songsInAlbum)
        {
            if (!Config.Album)
                return;
            
            foreach (var albumThatDoesNotNeedScanning in songsInAlbum.Where(entry => entry.Value.Length == 1)
                .Select(entry => entry.Value[0]))
            {
                if (ScanResults[albumThatDoesNotNeedScanning].Track is not null)
                    ScanResults[albumThatDoesNotNeedScanning].Album = ScanResults[albumThatDoesNotNeedScanning].Track;
            }
        }
    }
}