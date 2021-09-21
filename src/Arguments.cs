using System;
using System.Collections.Generic;
using System.Collections.Specialized;
using System.IO;
using System.Linq;
using System.Text;
using CommandLine;

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

        public SongsList(IEnumerable<string> songs)
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

        private string[] expandDirectories(IEnumerable<string> songs)
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
                    foreach (var subDirectory in subDirectories)
                    {
                        sc.AddRange(expandDirectories(new[] {subDirectory}));
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

    public class Options
    {
        [Value(0, Required = true)]
        public IEnumerable<string> Songs { get; set; }
        
        [Option('r', "track", Default = true, HelpText = "Calculate track gain.")]
        public bool Track { get; set;  }
        
        [Option('a', "album", Default = false, HelpText = "Calculate album gain (and track gain).")]
        public bool Album { get; set;  }
        
        [Option('k', "noclip", Default = false, HelpText = "Lower track/album gain to avoid clipping (<= -1 dBTP).")]
        public bool Noclip { get; set; }
        
        [Option('q', "quiet", Default = false, HelpText = "Don't print scanning status messages.")]
        public bool Quiet { get; set; }
        
        [Option('d', "pregain", Default = 0.0, HelpText = "Apply n dB/LU pre-gain value.")]
        public double _pregainFloat { get; set; }
        public Decibel Pregain { get; set; }
        
        [Option('K', "maxtpl", Default = -1.0,HelpText = "Avoid clipping; max. true peak level = n dBTP.")]
        public double _maxTruePeakLevelFloat { get; set; }
        public Decibel MaxTruePeakLevel { get; set; }

        public override string ToString()
        {
            return $"Songs: {String.Join(", ", this.Songs)}\n" +
                   $"Track: {this.Track}\n" +
                   $"Album: {this.Album}\n" +
                   $"Clipping prevention: {this.Noclip}\n" +
                   $"Quiet: {this.Quiet}\n" +
                   $"Pregain: {this.Pregain}\n" +
                   $"Max True Peak Level: {this.MaxTruePeakLevel}";
        }
    }
}