# Summary
This script takes a user's input text and instructs a Viam machine containing a robotic arm to draw the text.

![IMG_9657](https://github.com/user-attachments/assets/7ee0824a-fba0-4d1a-bc94-e8441774a8b8)

# Instructions

1. Make sure you have the Viam visualizer tool running, as per instructions at https://github.com/viam-labs/motion-tools
2. Create a Viam machine on app.viam.com, and [configure it with an arm](https://docs.viam.com/operate/mobility/move-arm/configure-arm/)
3. Add the Viam machine to the visualizer tool by clicking the small icon on the bottom right (see below for how to fill it out)

<img width="1334" alt="Screenshot 2025-05-30 at 11 38 44 PM" src="https://github.com/user-attachments/assets/2a0a0e5f-fc9b-4338-8b74-258a3f612334" />

The "Host" is the "Remote address" from the machine's "Live"/"Offline" dropdown on app.viam.com, and the "Part ID" is the "Part ID" from the same dropdown:

<img width="655" alt="Screenshot 2025-05-30 at 11 39 25 PM" src="https://github.com/user-attachments/assets/e22634b7-8979-46f8-acb7-76ddd1c1112d" />

The "API Key ID" and "API Key" can be found at the bottom of the "Organization Settings" page on app.viam.com.

The "Signaling Address" is "https://app.viam.com:443".

4. Configure the Viam machine with "Generic" components of model "fake" for any obstacles around the arm, such as the table the arm is mounted on. You can also configure "virtual walls" to constrain the arm's motion. Confirm the obstacles look as desired in the visualizer.

Example JSON configuration:

```
{
  "components": [
    {
      "name": "lite6-arm",
      "api": "rdk:component:arm",
      "model": "viam:ufactory:lite6",
      "attributes": {
        "host": "192.168.1.187",
        "speed_degs_per_sec": 10
      },
      "frame": {
        "parent": "world"
      }
    },
    {
      "name": "table",
      "api": "rdk:component:generic",
      "model": "rdk:builtin:fake",
      "attributes": {},
      "frame": {
        "parent": "world",
        "translation": {
          "x": 0,
          "y": 0,
          "z": 0
        },
        "orientation": {
          "type": "ov_degrees",
          "value": {
            "x": 0,
            "y": 0,
            "z": 1,
            "th": 0
          }
        },
        "geometry": {
          "type": "box",
          "x": 1000,
          "y": 1000,
          "z": 3
        }
      }
    },
    {
      "name": "fake-ceiling",
      "api": "rdk:component:generic",
      "model": "rdk:builtin:fake",
      "attributes": {},
      "frame": {
        "parent": "world",
        "translation": {
          "x": 0,
          "y": 0,
          "z": 650
        },
        "orientation": {
          "type": "ov_degrees",
          "value": {
            "x": 0,
            "y": 0,
            "z": 1,
            "th": 0
          }
        },
        "geometry": {
          "type": "box",
          "x": 1000,
          "y": 1000,
          "z": 5
        }
      }
    },
    {
      "name": "fake-back-wall",
      "api": "rdk:component:generic",
      "model": "rdk:builtin:fake",
      "attributes": {},
      "frame": {
        "parent": "world",
        "translation": {
          "x": 0,
          "y": -150,
          "z": 325
        },
        "orientation": {
          "type": "ov_degrees",
          "value": {
            "x": 0,
            "y": 0,
            "z": 1,
            "th": 0
          }
        },
        "geometry": {
          "type": "box",
          "x": 1000,
          "y": 5,
          "z": 650
        }
      }
    }
  ],
  "modules": [
    {
      "type": "registry",
      "name": "viam_ufactory",
      "module_id": "viam:ufactory",
      "version": "latest"
    }
  ]
}
```

Visualization for this config:
<img width="1462" alt="Screenshot 2025-05-30 at 11 36 34 PM" src="https://github.com/user-attachments/assets/fda27845-85dc-484b-bf6d-84ae8e4050c6" />

5. Clone this repo
6. Edit the "Viam machine info" [variables](https://github.com/EshaMaharishi/motion_demo/blob/49b00f74d89020f1a0109f4329110f33435f0e86/draw_points.go#L25-L28) in draw_points.go

- VIAM_MACHINE_IP_ADDRESS is the "Remote address" from the machine's "Live"/"Offline" dropdown on app.viam.com, as above
- API_KEY_ID and API_KEY are on your "Organization Settings" page on app.viam.com, as above
- ARM_RESOURCE_NAME is the name you gave your Viam machine's Arm component

7. Attach a pen or marker to the end of the arm, and place a notebook or whiteboard under the arm. Move the arm so that the pen/marker is touching the paper/whiteboard vertically, in the position you want it to write from. See the picture at the top for an example. To make sure the arm is in a comfortable position, make sure the joint positions are all between -180 and 180:
   
<img width="1128" alt="Screenshot 2025-05-30 at 11 50 29 PM" src="https://github.com/user-attachments/assets/a251f714-9401-4507-86d7-4c42c954e69c" />

8. Edit the "End effector" [variables](https://github.com/EshaMaharishi/motion_demo/blob/49b00f74d89020f1a0109f4329110f33435f0e86/draw_points.go#L35-L42) in draw_points.go to reflect the arm's current position. You can go to your Arm component's "Test" section and copy the Cartesian position and Orientation Vector values (under "Move to position") into the corresponding variables. Make sure to click "Current position" first. You can leave the ADDITIONAL_BUFFER_Z variable as-is.

<img width="1031" alt="Screenshot 2025-05-30 at 11 47 34 PM" src="https://github.com/user-attachments/assets/3bd55fec-0455-411b-92cd-bebb489617fd" />

9. From the root of this repo, run `script.sh`
10. At the prompt, enter a short (4 or 5 letter) word and hit enter. (The arm may not be able to reach far enough to write a longer word.)
11. The script will generate poses for the end effector to go to and send them to the visualizer. If the poses look reasonable, type 'y' to continue.

<img width="1460" alt="Screenshot 2025-05-30 at 11 54 22 PM" src="https://github.com/user-attachments/assets/1f25b8b8-4065-480f-a5bc-d39734d8bfa0" />

12. The script will create a client to your Viam machine and use the Motion service to tell the arm to go to the poses.

If you get an error saying the arm would collide with the table, try changing the END_EFFECTOR_Z_WHEN_PEN_TOUCHING_PAPER variable to be higher and see if you can trace the words in the air first, then gradually bring the value back down.
