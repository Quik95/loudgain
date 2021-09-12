#!/usr/bin/env python3

from dataclasses import dataclass
import os
import subprocess
import json
from typing import List, Tuple
import argparse
import pathlib
import math
import multiprocessing

parser = argparse.ArgumentParser()
parser.add_argument("first_dir", help="First folder containing audio files.")
parser.add_argument("second_dir", help="Second folder containing audio files.")


class Decibel(float):
    def __new__(cls, value=0) -> float:
        return float.__new__(cls, value)


class LoudnessUnit(float):
    def __new__(cls, value=0) -> float:
        return float.__new__(cls, value)


@dataclass
class ReplaygainValues:
    filename: str
    track_gain: Decibel
    track_peak: float
    track_range: Decibel
    reference_loudness: LoudnessUnit


def run():
    args = vars(parser.parse_args())

    songs_first_folder = get_songs_from_folder(args.get("first_dir", []))
    songs_second_folder = get_songs_from_folder(args.get("second_dir", []))

    songs_intersection = get_songs_intersection(songs_first_folder, songs_second_folder)

    with multiprocessing.Pool() as pool:
        results = pool.map(runner, songs_intersection)

        track_diff = get_diff([(x[0][0], x[1][0]) for x in results])
        album_diff = get_diff([(x[0][1], x[1][1]) for x in results])

        print("Track:")
        print(f"Gain difference: {track_diff[0][0]}: {track_diff[0][1]:.2f} dB")
        print(f"Peak difference: {track_diff[1][0]}: {track_diff[1][1]:.2f}")
        print(f"Range difference: {track_diff[2][0]}: {track_diff[2][1]:.2f} dB")

        print("\nAlbum:")
        print(f"Gain difference: {album_diff[0][0]}: {album_diff[0][1]:.2f} dB")
        print(f"Peak difference: {album_diff[1][0]}: {album_diff[1][1]:.2f}")
        print(f"Range difference: {album_diff[2][0]}: {album_diff[2][1]:.2f} dB")


def get_diff(
    results: List[Tuple[ReplaygainValues, ReplaygainValues]]
) -> Tuple[Tuple[str, float], Tuple[str, float], Tuple[str, float]]:
    max_gain_diff, max_gain_diff_file = -math.inf, ""
    max_peak_diff, max_peak_diff_file = -math.inf, ""
    max_range_diff, max_range_diff_file = -math.inf, ""

    for values_one, values_two in results:
        gain_diff, peak_diff, range_diff = get_value_difference(values_one, values_two)
        if gain_diff > max_gain_diff:
            max_gain_diff, max_gain_diff_file = gain_diff, values_one.filename
        if peak_diff > max_peak_diff:
            max_peak_diff, max_peak_diff_file = peak_diff, values_one.filename
        if range_diff > max_range_diff:
            max_range_diff, max_range_diff_file = range_diff, values_one.filename

    return (
        (max_gain_diff_file, max_gain_diff),
        (max_peak_diff_file, max_peak_diff),
        (max_range_diff_file, max_range_diff),
    )


def get_songs_intersection(
    list_one: List[str], list_two: List[str]
) -> List[Tuple[str, str]]:
    song_names_one = [pathlib.PurePath(song).stem for song in list_one]
    song_names_two = [pathlib.PurePath(song).stem for song in list_two]
    song_names_intersection = set.intersection(set(song_names_one), set(song_names_two))

    songs_list_one = [
        song
        for song in list_one
        if pathlib.PurePath(song).stem in song_names_intersection
    ]
    songs_list_two = [
        song
        for song in list_two
        if pathlib.PurePath(song).stem in song_names_intersection
    ]

    return list(zip(sorted(songs_list_one), sorted(songs_list_two)))


def get_songs_from_folder(filepath: str) -> List[str]:
    allowed_extension = set(
        [
            ".aiff",
            ".aif",
            ".aifc",
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
        ]
    )

    songs = []
    for root, _, files in os.walk(filepath):
        for file in files:
            if os.path.splitext(file)[1] in allowed_extension:
                songs.append(os.path.join(root, file))
    return songs


def print_json(pairs: List[Tuple[str, str]]):
    res = {}

    for tag_name, tag_value in pairs:
        if tag_name.startswith("replaygain_"):
            if tag_value.endswith("dB"):
                res[tag_name] = Decibel(value=tag_value[:-3])
            elif tag_value.endswith("LU"):
                res[tag_name] = LoudnessUnit(value=tag_value[:-3])
            elif tag_value.endswith("LUFS"):
                res[tag_name] = LoudnessUnit(value=tag_value[:-5])
            elif tag_name.endswith("peak"):
                res[tag_name] = float(tag_value)
        else:
            res[tag_name] = tag_value
    return res


def ProbeReplaygainValues(filename: str) -> Tuple[ReplaygainValues, ReplaygainValues]:
    probe = subprocess.run(
        args=[
            "ffprobe",
            "-hide_banner",
            "-show_format",
            "-print_format",
            "json",
            filename,
        ],
        stdout=subprocess.PIPE,
        stderr=subprocess.DEVNULL,
    )

    data = json.loads(probe.stdout, object_pairs_hook=print_json)
    top_level = data.get("format", {})
    tags = top_level.get("tags", {})

    return (
        ReplaygainValues(
            top_level.get("filename"),
            tags.get("replaygain_track_gain"),
            tags.get("replaygain_track_peak"),
            tags.get("replaygain_track_range"),
            tags.get("replaygain_reference_loudness"),
        ),
        ReplaygainValues(
            top_level.get("filename"),
            tags.get("replaygain_album_gain"),
            tags.get("replaygain_album_peak"),
            tags.get("replaygain_album_range"),
            tags.get("replaygain_reference_loudness"),
        ),
    )


def get_value_difference(
    values_one: ReplaygainValues, values_two: ReplaygainValues
) -> Tuple[float, float, float]:
    return (
        abs(values_one.track_gain - values_two.track_gain),
        abs(values_one.track_peak - values_two.track_peak),
        abs(values_one.track_range - values_two.track_range),
    )


def runner(songs: Tuple[str, str]) -> Tuple[Tuple[ReplaygainValues, ReplaygainValues], Tuple[ReplaygainValues, ReplaygainValues]]:
    return (
        ProbeReplaygainValues(songs[0]),
        ProbeReplaygainValues(songs[1]),
    )


def main():
    run()


if __name__ == "__main__":
    main()
