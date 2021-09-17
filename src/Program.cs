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

            var res = new ScanResult(songs.Songs[0]);
            await ScanResult.TrackScan(songs.Songs[0]);

            Console.WriteLine(res);
        }
    }
}