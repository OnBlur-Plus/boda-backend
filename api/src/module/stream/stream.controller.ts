import { BadRequestException, Body, Controller, Get, Post } from '@nestjs/common';
import { StreamService } from './stream.service';
import { IsPublic } from '../auth/auth.guard';

@Controller('stream')
export class StreamController {
  constructor(private readonly streamService: StreamService) {}

  @Get()
  async getAllStreams() {
    return await this.streamService.findAll();
  }

  @Get(':streamKey')
  async getStream(streamKey: string) {
    return await this.streamService.findOne(streamKey);
  }

  @Post('test')
  @IsPublic()
  async uploadFile(@Body() body: { file: string }) {
    const { file } = body;

    if (!file) {
      throw new BadRequestException('File data or name is missing');
    }

    return {
      imageId: 'test-id',
      x: 0.1,
      y: 0.1,
      width: 0.2,
      height: 0.2,
      label: 0,
      score: 0.2,
    };
  }
}
