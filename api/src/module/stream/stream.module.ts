import { Module } from '@nestjs/common';
import { StreamService } from './stream.service';
import { TypeOrmModule } from '@nestjs/typeorm';
import { Stream } from 'src/module/stream/entities/stream.entity';

@Module({
  imports: [TypeOrmModule.forFeature([Stream])],
  providers: [StreamService],
  exports: [TypeOrmModule],
})
export class StreamModule {}
