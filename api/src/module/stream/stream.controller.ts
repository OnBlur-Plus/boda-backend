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

  @Post('/verify')
  @IsPublic()
  async verifyStream(@Body('streamKey') streamKey: string) {
    await this.streamService.startStream(streamKey);
    return { message: 'Stream Verified' };
  }
  @Post('/end')
  @IsPublic()
  async endStream(@Body('streamKey') streamKey: string) {
    await this.streamService.endStream(streamKey);
    return { message: 'Stream ended' };
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
