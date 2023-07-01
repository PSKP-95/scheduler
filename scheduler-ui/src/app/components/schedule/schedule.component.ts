import { Component } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { History, Schedule } from 'src/app/models/schedule.model';
import { ScheduleService } from 'src/app/services/schedule.service';

@Component({
  selector: 'app-schedule',
  templateUrl: './schedule.component.html',
  styleUrls: ['./schedule.component.css']
})
export class ScheduleComponent {
  id: string = '';
  schedule: Schedule = new Schedule;
  history: History[] = [];

  constructor(private scheduleService: ScheduleService, private route: ActivatedRoute) { }

  ngOnInit(): void {
    this.route.params.subscribe(params => {
      this.id = params['id'];
    });

    this.loadSchedule();
  }

  loadSchedule() {
    this.scheduleService.loadSchedule(this.id).subscribe(
      data => {
        this.schedule = data;
      }
    );
    this.scheduleService.loadHistory(this.id).subscribe(
      data => {
        this.history = data;
      }
    )
  }

  isAcceptable(start: Date, end: Date): boolean {
    let enddate = new Date(end);
    let startdate = new Date(start);
    if (((enddate.getTime() - startdate.getTime()) / 1000) < 5)
      return true;
    return false;
  }
}
