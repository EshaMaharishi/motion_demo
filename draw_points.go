package main

import (
    "bufio"
	"context"
    "fmt"
    "math"
	"os"
	"strconv"
	"strings"

    "github.com/golang/geo/r3"
    vizClient "github.com/viam-labs/motion-tools/client/client"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/robot/client"
	"go.viam.com/utils/rpc"
	"go.viam.com/rdk/components/arm"
    "go.viam.com/rdk/services/motion"
    "go.viam.com/rdk/referenceframe"
    "go.viam.com/rdk/spatialmath"
    "go.viam.com/rdk/motionplan"
)

// Viam machine info
var VIAM_MACHINE_IP_ADDRESS = "your string here"
var API_KEY_ID = "your string here"
var API_KEY = "your string here"
var ARM_RESOURCE_NAME = "your string here"

// Text info
var POSE_COLOR_1 = "#EF5350"
var POSE_COLOR_2 = "#377EB8"
var POSE_COLOR_3 = "#4DAF4A"

// End effector position when pen touching paper where you want to start writing from
var END_EFFECTOR_X_WHEN_PEN_TOUCHING_PAPER float64 = 53.81019226834298
var END_EFFECTOR_Y_WHEN_PEN_TOUCHING_PAPER float64 = 261.357484082354 
var END_EFFECTOR_Z_WHEN_PEN_TOUCHING_PAPER float64 = 58.03686307529803
var ADDITIONAL_BUFFER_Z float64 = 20 // The initial position will be this many mm above the paper
var END_EFFECTOR_OV_X_WHEN_PEN_TOUCHING_PAPER float64 = 0.12923694451632337
var END_EFFECTOR_OV_Y_WHEN_PEN_TOUCHING_PAPER float64 = 0.30222728998800824
var END_EFFECTOR_OV_Z_WHEN_PEN_TOUCHING_PAPER float64 = -0.9444344748888563
var END_EFFECTOR_OV_THETA_WHEN_PEN_TOUCHING_PAPER float64 = 66.85935203732203

func main() {
    vizClient.RemoveAllSpatialObjects()

    // Variables to save the points read from the file
    var points [][][]float64    // 3D slice: points[letterIndex][pointIndex][xyz]
    var currentLetterPoints [][]float64

    // Variables to save the poses to move through
	var end_effector_poses []spatialmath.Pose
	var end_effector_colors []string

    // Open the file where the points were written
	file, err := os.Open("./generate_points/points3d.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

    // Read in the points from the file
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		if lineNum == 0 {
			// Skip the header
			lineNum++
			continue
		}

        // When a visual indicator line is found, start a new letter group
        if strings.HasPrefix(line, "# --- Letter:") {
            if len(currentLetterPoints) > 0 {
                // Save the previous letter's points before starting a new group
                points = append(points, currentLetterPoints)
            }
            currentLetterPoints = [][]float64{} // reset for new letter
            fmt.Printf("Loading points for letter: %s\n", line)
            continue
        }

		fields := strings.Split(line, ",")

		if len(fields) != 3 {
            fmt.Printf("Skipping malformed line: %s\n", line)
			continue
		}


		point := make([]float64, 3)
		for i, field := range fields {
			val, err := strconv.ParseFloat(strings.TrimSpace(field), 64)
			if err != nil {
				fmt.Printf("Error parsing float: %v\n", err)
				continue
			}
			point[i] = val
		}

        if (point[0] == math.Floor(point[0]) && point[1] == math.Floor(point[1])) {
            print("Skipping anchor point\n")
            continue
        }

        currentLetterPoints = append(currentLetterPoints, point)
		lineNum++
	}
    if len(currentLetterPoints) > 0 { // Save the final letter's points
        points = append(points, currentLetterPoints)
    }
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	// Construct the list of Poses the arm should move through to write the text
    for _, letter := range points {
        print("Generating poses for new letter\n")
        for i, pt := range letter {
            x, y, z := pt[0], pt[1], pt[2]

            // Always start a new letter at a pose above (positive z) its first point
            if (i == 0) {
                liftedPose := spatialmath.NewPose(
                    r3.Vector{X: x + END_EFFECTOR_X_WHEN_PEN_TOUCHING_PAPER, Y: y + END_EFFECTOR_Y_WHEN_PEN_TOUCHING_PAPER, Z: z + END_EFFECTOR_Z_WHEN_PEN_TOUCHING_PAPER + ADDITIONAL_BUFFER_Z},
                    &spatialmath.OrientationVectorDegrees{
                        OX: END_EFFECTOR_OV_X_WHEN_PEN_TOUCHING_PAPER,
                        OY: END_EFFECTOR_OV_Y_WHEN_PEN_TOUCHING_PAPER,
                        OZ: END_EFFECTOR_OV_Z_WHEN_PEN_TOUCHING_PAPER,
                        Theta: END_EFFECTOR_OV_THETA_WHEN_PEN_TOUCHING_PAPER,
                    },
                )
                end_effector_poses = append(end_effector_poses, liftedPose)
                end_effector_colors = append(end_effector_colors, POSE_COLOR_2)
            }

            // Add the current point for the letter as a pose
            end_effector_pose := spatialmath.NewPose(
                r3.Vector{X: x + END_EFFECTOR_X_WHEN_PEN_TOUCHING_PAPER, Y: y + END_EFFECTOR_Y_WHEN_PEN_TOUCHING_PAPER, Z: z + END_EFFECTOR_Z_WHEN_PEN_TOUCHING_PAPER},
                &spatialmath.OrientationVectorDegrees{
                    OX: END_EFFECTOR_OV_X_WHEN_PEN_TOUCHING_PAPER,
                    OY: END_EFFECTOR_OV_Y_WHEN_PEN_TOUCHING_PAPER,
                    OZ: END_EFFECTOR_OV_Z_WHEN_PEN_TOUCHING_PAPER,
                    Theta: END_EFFECTOR_OV_THETA_WHEN_PEN_TOUCHING_PAPER,
                },
            )
            end_effector_poses = append(end_effector_poses, end_effector_pose)
            if (i == len(letter) - 1) {
                end_effector_colors = append(end_effector_colors, POSE_COLOR_3)
            } else {
                end_effector_colors = append(end_effector_colors, POSE_COLOR_1)
            }

            // Always end a letter at a pose above (positive z) its last point
            if (i == len(letter) - 1) {
                liftedPose := spatialmath.NewPose(
                    r3.Vector{X: x + END_EFFECTOR_X_WHEN_PEN_TOUCHING_PAPER, Y: y + END_EFFECTOR_Y_WHEN_PEN_TOUCHING_PAPER, Z: z + END_EFFECTOR_Z_WHEN_PEN_TOUCHING_PAPER + ADDITIONAL_BUFFER_Z},
                    &spatialmath.OrientationVectorDegrees{
                        OX: END_EFFECTOR_OV_X_WHEN_PEN_TOUCHING_PAPER,
                        OY: END_EFFECTOR_OV_Y_WHEN_PEN_TOUCHING_PAPER,
                        OZ: END_EFFECTOR_OV_Z_WHEN_PEN_TOUCHING_PAPER,
                        Theta: END_EFFECTOR_OV_THETA_WHEN_PEN_TOUCHING_PAPER,
                    },
                )
                end_effector_poses = append(end_effector_poses, liftedPose)
                end_effector_colors = append(end_effector_colors, POSE_COLOR_2)
            }
        }
    }

    // Send the poses to the visualizer
	err = vizClient.DrawPoses(end_effector_poses, end_effector_colors, true)
	if err != nil {
        panic(err)
	}

    // Ask if the user wants to proceed with moving the arm through the poses
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Done generating poses! Look at the visualizer output. Do you want to continue to move the arm to the end_effector_poses? (y/n): ")
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err)
		os.Exit(1)
	}
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "y" || input == "yes" {
		fmt.Println("Continuing...")
	} else {
		fmt.Println("Exiting.")
		os.Exit(0)
	}

    // Connect to the Viam machine
    logger := logging.NewDebugLogger("client")
    machine, err := client.New(
        context.Background(),
        VIAM_MACHINE_IP_ADDRESS,
        logger,
        client.WithDialOptions(rpc.WithEntityCredentials( 
            API_KEY_ID,
            rpc.Credentials{
                Type:    rpc.CredentialsTypeAPIKey, 
                Payload: API_KEY,
            })),
    )
    if err != nil {
        logger.Fatal(err)
    }
    defer machine.Close(context.Background())
    
    // Get builtin motion service
    motionService, err := motion.FromRobot(machine, "builtin")
    if err != nil {
       logger.Fatal(err)
    }

    // Move arm to an initial pose.
    fmt.Print("Trying to move to initial pose: ")
    initialPose := spatialmath.NewPose(
        r3.Vector{X: END_EFFECTOR_X_WHEN_PEN_TOUCHING_PAPER, Y: END_EFFECTOR_Y_WHEN_PEN_TOUCHING_PAPER, Z: END_EFFECTOR_Z_WHEN_PEN_TOUCHING_PAPER + ADDITIONAL_BUFFER_Z},
        &spatialmath.OrientationVectorDegrees{
            OX: END_EFFECTOR_OV_X_WHEN_PEN_TOUCHING_PAPER,
            OY: END_EFFECTOR_OV_Y_WHEN_PEN_TOUCHING_PAPER,
            OZ: END_EFFECTOR_OV_Z_WHEN_PEN_TOUCHING_PAPER,
            Theta: END_EFFECTOR_OV_THETA_WHEN_PEN_TOUCHING_PAPER,
        },
    )
    initialPoseInFrame := referenceframe.NewPoseInFrame("world", initialPose)
    moved, err := motionService.Move(context.Background(), motion.MoveReq{
       ComponentName: arm.Named(ARM_RESOURCE_NAME),
       Destination:   initialPoseInFrame,
    })
    if err != nil {
       logger.Fatal(err)
    }
    logger.Info("moved", moved)

   myConstraints := &motionplan.Constraints{
      OrientationConstraint: []motionplan.OrientationConstraint{{1.0}},
   }

    // Move arm to the poses for tracing out the word
	for i, pose := range end_effector_poses {
        fmt.Print("Trying to move to pose: ")
		fmt.Print(pose)
        fmt.Println()

        destinationPoseInFrame := referenceframe.NewPoseInFrame("world", end_effector_poses[i])

        moved, err := motionService.Move(context.Background(), motion.MoveReq{
           ComponentName: arm.Named(ARM_RESOURCE_NAME),
           Destination:   destinationPoseInFrame,
           Constraints: myConstraints,
        })
        if err != nil {
           logger.Fatal(err)
        }
        logger.Info("moved", moved)
	}
}
