import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { History, Schedule, ScheduleHistoryResponse } from '../models/schedule.model';
import { Observable } from 'rxjs';
import {map} from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class ScheduleService {
  private scheduleUrl = "/api/schedule";
  private scheduleHistoryUrl = '/api/schedule/';
  schedule: Schedule;

  constructor(private http: HttpClient) { }

  loadSchedule(id: string): Observable<Schedule> {
    return this.http.get<any>(`${this.scheduleUrl}/${id}`)
      .pipe(
        map(value => this.schedule = value)
      );
  }

  loadHistory(id: string, size: number, page: number): Observable<ScheduleHistoryResponse> {
    return this.http.get<any>(`${this.scheduleHistoryUrl}/${id}/history?page=${page}&size=${size}`)
  }
}
