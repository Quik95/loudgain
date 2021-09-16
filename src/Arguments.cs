using System;

namespace loudgain
{
    class SongsList
    {
        public string[] Songs { get; }

        public SongsList(string[] songs)
        {
            var expanded = expandDirectories(songs);

            var badSong = CheckExtensions(expanded);
            if (badSong is not null)
            {
                Console.WriteLine(badSong);
                Environment.Exit(1);
            }

            this.Songs = expanded;
        }

        private static string? CheckExtensions(string[] songs)
        {
            var allowedExtensions = new System.Collections.Generic.HashSet<string>(
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

            foreach (var song in songs)
            {
                var extension = System.IO.Path.GetExtension(song);
                if (!allowedExtensions.Contains(extension))
                {
                    return $"invalid extension: {song}";
                }

                if (!System.IO.File.Exists(song))
                {
                    return $"file does not exits: {song}";
                }
            }

            return null;
        }

        private string[] expandDirectories(string[] songs)
        {
            var sc = new System.Collections.Specialized.StringCollection();

            foreach (var song in songs)
            {
                if (!System.IO.Directory.Exists(song))
                {
                    sc.Add(song);
                }

                var dirContents = System.IO.Directory.GetFiles(song);
                sc.AddRange(dirContents);

                var subDirectories = System.IO.Directory.GetDirectories(song);
                if (subDirectories.Length > 0)
                {
                    foreach (var subdir in subDirectories)
                    {
                        sc.AddRange(expandDirectories(new[] { subdir }));
                    }
                }
            }

            var res = new string[sc.Count];
            sc.CopyTo(res, 0);

            return res;
        }

        public override string ToString()
        {
            var sb = new System.Text.StringBuilder();

            foreach (var song in this.Songs)
            {
                sb.AppendLine(song);
            }

            return sb.ToString();
        }
    }
}