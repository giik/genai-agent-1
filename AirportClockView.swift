import SwiftUI

struct AirportClockView: View {
    @State private var currentTime = Date()

    let timer = Timer.publish(every: 1, on: .main, in: .common).autoconnect()

    var body: some View {
        ZStack {
            Color.black.edgesIgnoringSafeArea(.all)
            VStack {
                Text(timeString())
                    .font(.system(size: 60, weight: .bold, design: .monospaced))
                    .foregroundColor(.white)
            }
        }
        .onReceive(timer) { input in
            currentTime = input
        }
    }

    func timeString() -> String {
        let formatter = DateFormatter()
        formatter.dateFormat = "HH:mm:ss"
        return formatter.string(from: currentTime)
    }
}

struct AirportClockView_Previews: PreviewProvider {
    static var previews: some View {
        AirportClockView()
    }
}
