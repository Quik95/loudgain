using System;

namespace loudgain
{
    class Program
    {
        static void Main(string[] args)
        {
            var songs = new SongsList(args);
            Console.Write(songs);
        }
    }
}