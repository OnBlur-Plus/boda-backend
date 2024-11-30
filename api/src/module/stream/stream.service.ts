import { HttpException, HttpStatus, Injectable, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Stream, StreamStatus } from 'src/module/stream/entities/stream.entity';
import { CreateStreamDto } from './dto/create-stream.dto';
import { UpdateStreamDto } from './dto/update-stream.dto';

@Injectable()
export class StreamService {
  constructor(
    @InjectRepository(Stream)
    private readonly streamRepository: Repository<Stream>,
  ) {}

  async findAll() {
    return await this.streamRepository.find();
  }

  async findStream(streamKey: string) {
    return await this.streamRepository.findOne({ where: { streamKey } });
  }

  async getStreamURL(streamKey: string) {
    const stream = await this.streamRepository.findOne({
      where: { streamKey },
      relations: ['user'],
    });

    if (!stream) {
      throw new HttpException('Stream not found', HttpStatus.NOT_FOUND);
    }

    return true;
  }

  async createStream(createStreamDto: CreateStreamDto) {
    return await this.streamRepository.save(createStreamDto);
  }

  async updateStream(streamKey: string, updateStreamDto: UpdateStreamDto) {
    return await this.streamRepository.update({ streamKey }, updateStreamDto);
  }

  async deleteStream(streamKey: string) {
    return await this.streamRepository.delete({ streamKey });
  }

  async startStream(streamKey: string) {
    const stream = await this.streamRepository.findOne({ where: { streamKey } });

    if (!stream) {
      return new NotFoundException('No Stream');
    }

    return await this.streamRepository.update(
      { streamKey },
      {
        status: StreamStatus.LIVE,
      },
    );
  }

  async endStream(streamKey: string) {
    const stream = await this.streamRepository.findOne({ where: { streamKey } });

    if (!stream) {
      return new NotFoundException('No Stream');
    }

    return await this.streamRepository.update(
      { streamKey },
      {
        status: StreamStatus.IDLE,
      },
    );
  }
}
