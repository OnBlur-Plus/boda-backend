import { Injectable } from '@nestjs/common';
import { CreateStreamDto } from './dto/create-stream.dto';
import { UpdateStreamDto } from './dto/update-stream.dto';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Stream } from 'src/module/stream/entities/stream.entity';

@Injectable()
export class StreamService {
  constructor(
    @InjectRepository(Stream)
    private readonly streamRepo: Repository<Stream>,
  ) {}
  create(createStreamDto: CreateStreamDto) {
    return 'This action adds a new stream';
  }

  findAll() {
    return `This action returns all stream`;
  }

  findOne(id: number) {
    return `This action returns a #${id} stream`;
  }

  update(id: number, updateStreamDto: UpdateStreamDto) {
    return `This action updates a #${id} stream`;
  }

  remove(id: number) {
    return `This action removes a #${id} stream`;
  }
}
