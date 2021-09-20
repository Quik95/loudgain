using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Linq;
using System.Threading.Tasks;
using File = TagLib.File;

namespace loudgain
{
    public static class Album
    {
        public static Dictionary<string, string[]> GetSongsInAlbum(List<string> songList)
        {
            var res = new Dictionary<string, string[]>();

            var songPlusAlbum = songList.Select(_getSongAlbum).ToArray();
            var uniqueAlbums = new HashSet<string>(
                songPlusAlbum.Where(tuple => tuple.Item2 is not null).Select(tuple => tuple.Item2)
                    .OfType<string>()
            );
            foreach (var album in uniqueAlbums)
            {
                res.Add(album, songPlusAlbum.Where(s => s.Item2 == album).Select(s => s.Item1).ToArray());
            }

            return res;
        }

        private static Tuple<string, string?> _getSongAlbum(string song)
        {
            try
            {
                var file = File.Create(song);
                return new Tuple<string, string?>(song, file.Tag.Album);
            }
            catch
            {
                Console.Error.WriteLine($"song: {song} appears to be corrupt");
                return new Tuple<string, string?>(song, null);
            }
        }

        private static bool _checkSameExtension(string[] songs)
        {
            var ext = Path.GetExtension(songs[0]);
            return songs.All(song => Path.GetExtension(song) == ext);
        }

        public static async Task<ReplaygainValues?> ScanAlbum(string[] albumSongs)
        {
            if (!_checkSameExtension(albumSongs))
            {
                Program.MasterProgressBar.WriteLine("All songs in an album must have the same extension:");
                foreach (var song in albumSongs)
                {
                    Program.MasterProgressBar.WriteLine(Path.GetFileName(song));
                }

                return null;
            }

            var joinedSongs = await _joinAlbumSongs(albumSongs);
            if (joinedSongs is null)
                return null;
            var res = await ScanResult.TrackScan(joinedSongs);
            return res;
        }

        private static async Task<string?> _joinAlbumSongs(string[] songs)
        {
            var ffmpegConcatInputFile = await _createFFmpegConcatFile(songs);
            if (ffmpegConcatInputFile is null)
                return null;
            var outFile = new TempFile(Path.ChangeExtension(Path.Combine(Path.GetTempPath(), Path.GetRandomFileName()),
                Path.GetExtension(songs[0]))).FileName;

            var args = $"-hide_banner -y -f concat -safe 0 -i {ffmpegConcatInputFile} -c copy {outFile}";

            using var process = new Process();
            process.StartInfo.UseShellExecute = false;
            process.StartInfo.FileName = "ffmpeg";
            process.StartInfo.Arguments = args;
            process.StartInfo.CreateNoWindow = false;
            process.StartInfo.RedirectStandardOutput = true;
            process.StartInfo.RedirectStandardError = true;
            process.Start();

            var potentialError = process.StandardError.ReadToEndAsync();
            await process.WaitForExitAsync();

            if (process.ExitCode != 0)
            {
                Program.MasterProgressBar.WriteErrorLine($"failed to concatenate album songs: {await potentialError}");
                return null;
            }

            return outFile;
        }

        private static async Task<string?> _createFFmpegConcatFile(string[] songs)
        {
            try
            {
                var tmp = new TempFile();
                await using var file = new StreamWriter(tmp.FileName);
                foreach (var song in songs)
                {
                    var quotesEscaped = _escapeQuotes(song);
                    await file.WriteLineAsync($"file '{quotesEscaped}'");
                }

                return tmp.FileName;
            }
            catch
            {
                return null;
            }
        }

        private static string _escapeQuotes(string song)
        {
            return song.Replace(@"'", @"'\''");
        }
    }

    class TempFile
    {
        public string FileName { get; }

        private void _disposeFile(object? sender, ConsoleCancelEventArgs consoleCancelEventArgs)
        {
            System.IO.File.Delete(this.FileName);
        }

        private void _disposeFile()
        {
            System.IO.File.Delete(this.FileName);
        }

        public TempFile(string filename)
        {
            this.FileName = filename;
            Console.CancelKeyPress += this._disposeFile;
        }

        public TempFile()
        {
            this.FileName = Path.GetTempFileName();
            Console.CancelKeyPress += this._disposeFile;
        }

        ~TempFile()
        {
            Console.CancelKeyPress -= this._disposeFile;
            this._disposeFile();
        }
    }
}