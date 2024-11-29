import { Controller, Get, Post, UploadedFile, UseInterceptors } from '@nestjs/common';
import { StreamService } from './stream.service';
import { FileInterceptor } from '@nestjs/platform-express';
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
  @UseInterceptors(FileInterceptor('file'))
  async getBoundingBox(@UploadedFile() file: Express.Multer.File) {
    console.log(file);

    return {
      imageId: 'test-id',
      x1: 0.1,
      y1: 0.1,
      x2: 0.2,
      y2: 0.2,
      label: 0,
      score: 0.2,
    };
  }
}
