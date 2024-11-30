import { Entity, Column, PrimaryGeneratedColumn, UpdateDateColumn, CreateDateColumn } from 'typeorm';

export enum StreamStatus {
  IDLE = 'IDLE',
  LIVE = 'LIVE',
}

@Entity()
export class Stream {
  @PrimaryGeneratedColumn('uuid', { name: 'stream_key' })
  streamKey: string;

  @Column()
  title: string;

  @Column({ name: 'sub_title' })
  subTitle: string;

  @Column({ name: 'thumbnail_url', nullable: true })
  thumbnailUrl: string;

  @Column({ type: 'enum', enum: StreamStatus })
  status: StreamStatus;

  @CreateDateColumn({ name: 'created_at' })
  createdAt: Date;

  @UpdateDateColumn({ name: 'updated_at' })
  updatedAt: Date;
}
