import {
  Entity,
  Column,
  PrimaryGeneratedColumn,
  DeleteDateColumn,
  UpdateDateColumn,
  CreateDateColumn,
  OneToOne,
} from 'typeorm';

@Entity()
export class Stream {
  @PrimaryGeneratedColumn('uuid', { name: 'stream_key' })
  streamKey: string;

  test: string;
}
