import { BadRequestException, Body, Controller, Delete, Get, Param, Post } from '@nestjs/common';
import { StreamService } from './stream.service';
import { IsPublic } from '../auth/auth.guard';
import { CreateStreamDto } from './dto/create-stream.dto';
import { UpdateStreamDto } from './dto/update-stream.dto';

@Controller('stream')
export class StreamController {
  constructor(private readonly streamService: StreamService) {}

  @Get()
  async getAllStreams() {
    return await this.streamService.findAll();
  }

  @Get(':streamKey')
  async getStream(streamKey: string) {
    return await this.streamService.findStream(streamKey);
  }

  @Post()
  async createStream(@Body() createStreamDto: CreateStreamDto) {
    return await this.streamService.createStream(createStreamDto);
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

  @Post(':streamKey')
  async updateStream(@Param('streamKey') streamKey: string, @Body() updateStreamDto: UpdateStreamDto) {
    return await this.streamService.updateStream(streamKey, updateStreamDto);
  }

  @Delete(':streamKey')
  async deleteStream(@Param('streamKey') streamKey: string) {
    return await this.streamService.deleteStream(streamKey);
  }
}
