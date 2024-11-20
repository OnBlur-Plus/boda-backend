from watchdog.observers import Observer
from watchdog.events import FileDeletedEvent, FileSystemEventHandler
from asyncio import AbstractEventLoop
import os
import asyncio
import random
from manifest_parser import is_master_manifest, get_last_segment_and_start_timestamp

UTC_TIME_FMT = "%Y-%m-%dT%H:%M:%S.%fZ"
PROGRAM_TIME_KEYWORD = "#EXT-X-PROGRAM-DATE-TIME:"
DURATION_KEYWORD = "EXTINF:"
SEGMENT_SUFFIX = ".ts"


class FileChangeAsyncHandler(FileSystemEventHandler):
    def __init__(self, target_file, queue, loop: AbstractEventLoop, observer):
        self.target_file = target_file
        self.observer = observer
        self.loop = loop
        self.queue = queue

    def on_modified(self, event):
        if event.src_path == self.target_file:
            self.loop.call_soon_threadsafe(
                asyncio.create_task, on_modified_async(self.target_file, self.queue)
            )

    def on_deleted(self, event: FileDeletedEvent):
        self.observer.stop()


async def on_modified_async(target_file, queue):
    # parse m3u8
    segment_file = None
    with open(target_file, "r") as f:
        lines = f.read().splitlines()
        if is_master_manifest(lines):
            pass
        else:
            segment_file, starting_time, duration_sec = (
                get_last_segment_and_start_timestamp(lines)
            )

    if segment_file is None:
        return

    task = asyncio.create_task(network_request(segment_file))
    await queue.put(task)


async def run_observer(target_file, queue, loop):
    directory = os.path.dirname(target_file) or "."

    observer = Observer()
    event_handler = FileChangeAsyncHandler(target_file, queue, loop, observer)
    observer.schedule(event_handler, path=directory, recursive=False)

    observer.start()
    print(f"파일 감시 중: {target_file}")

    try:
        while observer.is_alive():
            await asyncio.sleep(1)
    except asyncio.CancelledError:
        print("Observer 중단 요청")
    finally:
        observer.stop()
        observer.join()
        print("Observer 종료")


async def monitor_async(target_file):
    queue = asyncio.Queue()
    loop = asyncio.get_running_loop()
    only_task = asyncio.create_task(on_modified_async(target_file, queue))
    producer_task = asyncio.create_task(run_observer(target_file, queue, loop))
    consumer_task = asyncio.create_task(consumer(queue))
    try:
        await only_task
        await producer_task
        await queue.join()
    except KeyboardInterrupt:
        producer_task.cancel()
    finally:
        consumer_task.cancel()


def monitor(target_file):
    asyncio.run(monitor_async(target_file))


async def network_request(segment_file):
    await asyncio.sleep(random.uniform(0.1, 1.0))

    return segment_file


async def consumer(queue):
    last_file = None
    while True:
        # 큐에서 다음 태스크를 가져옴
        task = await queue.get()
        try:
            # 태스크가 완료될 때까지 대기하고 결과 처리
            file = await task
            if file == last_file:
                continue
            print(f"소비자: {file}")
            last_file = file
        finally:
            queue.task_done()
