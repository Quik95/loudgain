using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Linq;
using System.Threading.Tasks;

namespace loudgain
{
    public class Album
    {
        public static Dictionary<string, string[]> GetSongsInAlbum(List<string> songList)
        {
            var res = new Dictionary<string, string[]>();

            var songPlusAlbum = songList.Select(_getSongAlbum).ToArray();
            var uniqueAlbums = new HashSet<string>(
                songPlusAlbum.Where(song => song.Item2 is not null).Select(songPlusAlbum => songPlusAlbum.Item2)
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
                var file = TagLib.File.Create(song);
                return new Tuple<string, string?>(song, file.Tag.Album);
            }
            catch
            {
                return new Tuple<string, string?>(song, null);
            }
        }

        public static async Task<ReplaygainValues?> ScanAlbum(string[] albumSongs)
        {
            if (albumSongs.Length == 1)
            {
                return null;
            }

            var joinedSongs = await _joinAlbumSongs(albumSongs);
            if (joinedSongs is null)
                return null;
            return await ScanResult.TrackScan(joinedSongs);
        }

        private static async Task<string?> _joinAlbumSongs(string[] songs)
        {
            var ffmpegConcatInputFile = await _createFFmpegConcatFile(songs);
            if (ffmpegConcatInputFile is null)
                return null;

            var outFile = Path.ChangeExtension(Path.Combine(Path.GetTempPath(), Path.GetRandomFileName()), Path.GetExtension(songs[0]));

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
                await Console.Error.WriteLineAsync($"failed to concatenate album songs: {await potentialError}");
                return null;
            }

            return outFile;
        }

        private static async Task<string?> _createFFmpegConcatFile(string[] songs)
        {
            try
            {
                var filename = Path.GetTempFileName();
                await using var file = new StreamWriter(filename);
                foreach (var song in songs)
                {
                    var quotesEscaped = _escapeQuotes(song);
                    await file.WriteLineAsync($"file '{quotesEscaped}'");
                }

                return filename;
            }
            catch (IOException e)
            {
                return null;
            }
        }

        private static string _escapeQuotes(string song)
        {
            return song.Replace(@"'", @"'\''");
        }
    }
}