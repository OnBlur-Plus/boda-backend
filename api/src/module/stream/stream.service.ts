import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Stream } from 'src/module/stream/entities/stream.entity';

@Injectable()
export class StreamService {
  constructor(
    @InjectRepository(Stream)
    private readonly streamRepository: Repository<Stream>,
  ) {}

  async findAll() {
    return await this.streamRepository.find();
  }

  async findOne(streamKey: string) {
    return await this.streamRepository.findOne({ where: { streamKey } });
  }
}
