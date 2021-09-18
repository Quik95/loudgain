using System;
using System.Threading.Tasks;

namespace loudgain
{
    class Program
    {
        static async Task Main(string[] args)
        {
            var songs = new SongsList(args);
            Console.Write(songs);

            foreach (var song in songs.Songs)
            {
                var res = await ScanResult.TrackScan(song);
                Console.WriteLine($"song: {song} => {res}");
            }
        }
    }
}