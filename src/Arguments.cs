using System;
using System.Collections.Generic;
using System.Collections.Specialized;
using System.IO;
using System.Linq;
using System.Text;

namespace loudgain
{
    class SongsList
    {
        private static readonly HashSet<string> AllowedExtensions = new HashSet<string>(
            new[]
            {
                ".aiff",
                ".aif",
                ".alfc",
                ".ape",
                ".apl",
                ".bwf",
                ".flac",
                ".mp3",
                ".mp4",
                ".m4a",
                ".m4b",
                ".m4p",
                ".m4r",
                ".mpc",
                ".ogg",
                ".tta",
                ".wma",
                ".wv",
            }
        );

        public List<string> Songs { get; }

        public SongsList(string[] songs)
        {
            var expanded = expandDirectories(songs);

            var allowedSongs = new List<string>(expanded.Where(CheckExtension));
            if (allowedSongs.Count == 0)
            {
                Console.WriteLine("No songs provided. Exiting...");
                Environment.Exit(-1);
            }

            this.Songs = allowedSongs;
        }

        private static bool CheckExtension(string song)
        {
            var extension = Path.GetExtension(song);
            if (!AllowedExtensions.Contains(extension))
                return false;

            if (!File.Exists(song))
                return false;

            return true;
        }

        private string[] expandDirectories(string[] songs)
        {
            var sc = new StringCollection();

            foreach (var song in songs)
            {
                if (File.Exists(song))
                {
                    sc.Add(song);
                    continue;
                }

                var dirContents = Directory.GetFiles(song);
                sc.AddRange(dirContents);

                var subDirectories = Directory.GetDirectories(song);
                if (subDirectories.Length > 0)
                {
                    foreach (var subdir in subDirectories)
                    {
                        sc.AddRange(expandDirectories(new[] {subdir}));
                    }
                }
            }

            var res = new string[sc.Count];
            sc.CopyTo(res, 0);

            return res;
        }

        public override string ToString()
        {
            var sb = new StringBuilder();

            foreach (var song in this.Songs)
            {
                sb.AppendLine(song);
            }

            return sb.ToString();
        }
    }
}