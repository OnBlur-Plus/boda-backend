import { Module } from '@nestjs/common';
import { StreamService } from './stream.service';
import { TypeOrmModule } from '@nestjs/typeorm';
import { Stream } from 'src/module/stream/entities/stream.entity';
import { StreamController } from './stream.controller';

@Module({
  imports: [TypeOrmModule.forFeature([Stream])],
  controllers: [StreamController],
  providers: [StreamService],
  exports: [TypeOrmModule, StreamService],
})
export class StreamModule {}
