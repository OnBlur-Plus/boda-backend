from watchdog.observers import Observer
from watchdog.events import PatternMatchingEventHandler
from concurrent.futures import ProcessPoolExecutor
import time
import os
from process_playlist import monitor
import sys


class FileCreationHandler(PatternMatchingEventHandler):
    def __init__(
        self,
        *,
        patterns: list[str] | None = None,
        ignore_patterns: list[str] | None = None,
        ignore_directories: bool = False,
        case_sensitive: bool = False,
        detector: dict,
    ):
        super().__init__(
            patterns=patterns,
            ignore_patterns=ignore_patterns,
            ignore_directories=ignore_directories,
            case_sensitive=case_sensitive,
        )
        self.detector = detector

    def __enter__(self):
        self.executer = ProcessPoolExecutor(max_workers=10)
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.executer.shutdown(wait=True)
        return False

    def on_modified(self, event):
        if not event.is_directory:
            file_path = event.src_path
            if file_path in self.detector:
                return
            future = self.executer.submit(monitor, file_path)
            future.add_done_callback(lambda f: self.detector.pop(file_path))
            self.detector[file_path] = future


if __name__ == "__main__":
    watch_directory = sys.argv[1] or "."
    detector = {}
    with FileCreationHandler(patterns=["*.m3u8"], detector=detector) as event_handler:
        observer = Observer()
        observer.schedule(event_handler, path=watch_directory, recursive=True)
        observer.start()
        print(f"디렉토리 감시 중: {os.path.abspath(watch_directory)}")
        try:
            while True:
                time.sleep(1)
        except KeyboardInterrupt:
            observer.stop()
        observer.join()
